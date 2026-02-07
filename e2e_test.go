//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/classi/persistence"
	"github.com/emiliopalmerini/quintaedizione.api/internal/classi/transports"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

const testDocRiferimento = "DND 2024"

var api *testAPI

type testAPI struct {
	container *tcpostgres.PostgresContainer
	db        *sqlx.DB
	server    *httptest.Server
}

func TestMain(m *testing.M) {
	if !isDockerAvailable() {
		log.Println("Docker is not available, skipping e2e tests")
		os.Exit(0)
	}

	var err error
	api, err = newTestAPI()
	if err != nil {
		log.Fatalf("failed to setup test API: %v", err)
	}

	code := m.Run()

	api.shutdown()
	os.Exit(code)
}

func isDockerAvailable() bool {
	return exec.Command("docker", "info").Run() == nil
}

func newTestAPI() (*testAPI, error) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("get mapped port: %w", err)
	}

	connStr := fmt.Sprintf("postgres://testuser:testpass@127.0.0.1:%s/testdb?sslmode=disable", port.Port())

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := runTestMigrations(db); err != nil {
		db.Close()
		container.Terminate(ctx)
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	repo := persistence.NewPostgresRepository(db)
	service := classi.NewService(repo, nil) // nil logger defaults to discard
	handler := transports.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})
	r.Mount("/v1/classi", handler.Routes())

	server := httptest.NewServer(r)

	return &testAPI{
		container: container,
		db:        db,
		server:    server,
	}, nil
}

func (a *testAPI) shutdown() {
	a.server.Close()
	a.db.Close()
	a.container.Terminate(context.Background())
}

func (a *testAPI) truncateTables(t *testing.T) {
	t.Helper()
	_, err := a.db.Exec("TRUNCATE TABLE sottoclassi, classi CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

func (a *testAPI) insertClasse(t *testing.T, id, nome string, dadoVita classi.TipoDiDado) {
	t.Helper()
	_, err := a.db.Exec(`
		INSERT INTO classi (id, nome, descrizione, documentazione_di_riferimento, dado_vita)
		VALUES ($1, $2, $3, $4, $5)
	`, id, nome, "Test description for "+nome, testDocRiferimento, string(dadoVita))
	if err != nil {
		t.Fatalf("failed to insert classe: %v", err)
	}
}

func (a *testAPI) insertSottoclasse(t *testing.T, id, nome, classeID string) {
	t.Helper()
	_, err := a.db.Exec(`
		INSERT INTO sottoclassi (id, nome, descrizione, documentazione_di_riferimento, id_classe_associata)
		VALUES ($1, $2, $3, $4, $5)
	`, id, nome, "Test description for "+nome, testDocRiferimento, classeID)
	if err != nil {
		t.Fatalf("failed to insert sottoclasse: %v", err)
	}
}

// get performs a GET request, asserts the status code, and decodes the
// JSON body into target (if non-nil). It returns the raw response for
// additional assertions.
func (a *testAPI) get(t *testing.T, path string, wantStatus int, target any) *http.Response {
	t.Helper()
	resp, err := http.Get(a.server.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("GET %s: expected Content-Type application/json, got %q", path, ct)
	}
	if resp.StatusCode != wantStatus {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s: expected status %d, got %d; body: %s", path, wantStatus, resp.StatusCode, body)
	}
	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			t.Fatalf("GET %s: failed to decode response: %v", path, err)
		}
	}
	return resp
}

func runTestMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "migrations")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func TestE2E_HealthCheck(t *testing.T) {
	var result map[string]string
	api.get(t, "/health", http.StatusOK, &result)

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result["status"])
	}
}

func TestE2E_ListClassi(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		api.truncateTables(t)

		var result classi.ListClassiResponse
		api.get(t, "/v1/classi", http.StatusOK, &result)

		if result.NumeroDiElementi != 0 {
			t.Errorf("expected 0 elements, got %d", result.NumeroDiElementi)
		}
	})

	// Seed data for subsequent subtests.
	api.truncateTables(t)
	api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
	api.insertClasse(t, "mago", "Mago", classi.D6)
	api.insertClasse(t, "guerriero", "Guerriero", classi.D10)

	t.Run("with data", func(t *testing.T) {
		var result classi.ListClassiResponse
		api.get(t, "/v1/classi", http.StatusOK, &result)

		if result.NumeroDiElementi != 3 {
			t.Errorf("expected 3 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Classi) != 3 {
			t.Errorf("expected 3 classi, got %d", len(result.Classi))
		}
	})

	t.Run("with name filter", func(t *testing.T) {
		var result classi.ListClassiResponse
		api.get(t, "/v1/classi?nome=bar", http.StatusOK, &result)

		if result.NumeroDiElementi != 1 {
			t.Errorf("expected 1 element, got %d", result.NumeroDiElementi)
		}
		if result.Classi[0].ID != "barbaro" {
			t.Errorf("expected 'barbaro', got '%s'", result.Classi[0].ID)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		var result classi.ListClassiResponse
		api.get(t, "/v1/classi?$limit=2&$offset=0", http.StatusOK, &result)

		if result.NumeroDiElementi != 3 {
			t.Errorf("expected 3 total elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Classi) != 2 {
			t.Errorf("expected 2 classi in page, got %d", len(result.Classi))
		}
	})

	t.Run("with desc sort", func(t *testing.T) {
		var result classi.ListClassiResponse
		api.get(t, "/v1/classi?sort=desc", http.StatusOK, &result)

		if result.Classi[0].Nome != "Mago" {
			t.Errorf("expected first classe 'Mago' (desc order), got '%s'", result.Classi[0].Nome)
		}
	})

	t.Run("with second page", func(t *testing.T) {
		var result classi.ListClassiResponse
		api.get(t, "/v1/classi?$limit=2&$offset=2", http.StatusOK, &result)

		if result.NumeroDiElementi != 3 {
			t.Errorf("expected 3 total elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Classi) != 1 {
			t.Errorf("expected 1 classe on second page, got %d", len(result.Classi))
		}
	})

	t.Run("with documentazione filter", func(t *testing.T) {
		var result classi.ListClassiResponse
		api.get(t, "/v1/classi?documentazione-di-riferimento=DND+2024", http.StatusOK, &result)

		if result.NumeroDiElementi != 3 {
			t.Errorf("expected 3 elements, got %d", result.NumeroDiElementi)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi?$limit=-1", http.StatusBadRequest, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
		if result.Errors[0].Code != "BAD_REQUEST" {
			t.Errorf("expected code 'BAD_REQUEST', got '%s'", result.Errors[0].Code)
		}
	})

	t.Run("limit exceeds max", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi?$limit=999", http.StatusBadRequest, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
	})

	t.Run("invalid offset", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi?$offset=-1", http.StatusBadRequest, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
	})

	t.Run("invalid sort", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi?sort=invalid", http.StatusBadRequest, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
	})
}

func TestE2E_GetClasse(t *testing.T) {
	api.truncateTables(t)
	api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
	api.insertSottoclasse(t, "berserker", "Berserker", "barbaro")
	api.insertSottoclasse(t, "totemico", "Totemico", "barbaro")

	t.Run("existing classe", func(t *testing.T) {
		var result classi.Classe
		api.get(t, "/v1/classi/barbaro", http.StatusOK, &result)

		if result.ID != "barbaro" {
			t.Errorf("expected id 'barbaro', got '%s'", result.ID)
		}
		if result.Nome != "Barbaro" {
			t.Errorf("expected nome 'Barbaro', got '%s'", result.Nome)
		}
		if result.DadoVita != classi.D12 {
			t.Errorf("expected dado_vita 'd12', got '%s'", result.DadoVita)
		}
		if len(result.ElencoSottoclassi) != 2 {
			t.Errorf("expected 2 sottoclassi references, got %d", len(result.ElencoSottoclassi))
		}
	})

	t.Run("non-existing classe", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi/nonexistent", http.StatusNotFound, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
		if result.Errors[0].Code != "NOT_FOUND" {
			t.Errorf("expected code 'NOT_FOUND', got '%s'", result.Errors[0].Code)
		}
	})
}

func TestE2E_ListSottoclassi(t *testing.T) {
	api.truncateTables(t)
	api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
	api.insertSottoclasse(t, "berserker", "Berserker", "barbaro")
	api.insertSottoclasse(t, "totemico", "Totemico", "barbaro")

	t.Run("list sottoclassi", func(t *testing.T) {
		var result classi.ListSottoclassiResponse
		api.get(t, "/v1/classi/barbaro/sotto-classi", http.StatusOK, &result)

		if result.NumeroDiElementi != 2 {
			t.Errorf("expected 2 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Sottoclassi) != 2 {
			t.Errorf("expected 2 sottoclassi, got %d", len(result.Sottoclassi))
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi/nonexistent/sotto-classi", http.StatusNotFound, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
		if result.Errors[0].Code != "NOT_FOUND" {
			t.Errorf("expected code 'NOT_FOUND', got '%s'", result.Errors[0].Code)
		}
	})

	t.Run("with name filter", func(t *testing.T) {
		var result classi.ListSottoclassiResponse
		api.get(t, "/v1/classi/barbaro/sotto-classi?nome=ber", http.StatusOK, &result)

		if result.NumeroDiElementi != 1 {
			t.Errorf("expected 1 element, got %d", result.NumeroDiElementi)
		}
		if result.Sottoclassi[0].ID != "berserker" {
			t.Errorf("expected 'berserker', got '%s'", result.Sottoclassi[0].ID)
		}
	})
}

func TestE2E_GetSottoclasse(t *testing.T) {
	api.truncateTables(t)
	api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
	api.insertSottoclasse(t, "berserker", "Berserker", "barbaro")

	t.Run("existing sottoclasse", func(t *testing.T) {
		var result classi.SottoClasse
		api.get(t, "/v1/classi/barbaro/sotto-classi/berserker", http.StatusOK, &result)

		if result.ID != "berserker" {
			t.Errorf("expected id 'berserker', got '%s'", result.ID)
		}
		if result.Nome != "Berserker" {
			t.Errorf("expected nome 'Berserker', got '%s'", result.Nome)
		}
		if result.IDClasseAssociata != "barbaro" {
			t.Errorf("expected id_classe_associata 'barbaro', got '%s'", result.IDClasseAssociata)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi/nonexistent/sotto-classi/berserker", http.StatusNotFound, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
		if result.Errors[0].Code != "NOT_FOUND" {
			t.Errorf("expected code 'NOT_FOUND', got '%s'", result.Errors[0].Code)
		}
	})

	t.Run("sottoclasse not found", func(t *testing.T) {
		var result shared.ErrorObject
		api.get(t, "/v1/classi/barbaro/sotto-classi/nonexistent", http.StatusNotFound, &result)

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
		if result.Errors[0].Code != "NOT_FOUND" {
			t.Errorf("expected code 'NOT_FOUND', got '%s'", result.Errors[0].Code)
		}
	})
}
