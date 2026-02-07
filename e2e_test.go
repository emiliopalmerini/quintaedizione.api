//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func skipIfDockerNotAvailable(t *testing.T) {
	t.Helper()
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker is not available, skipping e2e tests")
	}
}

type testAPI struct {
	container *tcpostgres.PostgresContainer
	db        *sqlx.DB
	server    *httptest.Server
}

func setupTestAPI(t *testing.T) *testAPI {
	t.Helper()
	skipIfDockerNotAvailable(t)
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
		t.Fatalf("failed to start postgres container: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get mapped port: %v", err)
	}

	connStr := fmt.Sprintf("postgres://testuser:testpass@127.0.0.1:%s/testdb?sslmode=disable", port.Port())

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := runTestMigrations(db); err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to run migrations: %v", err)
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
	}
}

func (api *testAPI) cleanup(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	api.server.Close()
	api.db.Close()
	api.container.Terminate(ctx)
}

func (api *testAPI) truncateTables(t *testing.T) {
	t.Helper()
	_, err := api.db.Exec("TRUNCATE TABLE sottoclassi, classi CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

func (api *testAPI) insertClasse(t *testing.T, id, nome string, dadoVita classi.TipoDiDado) {
	t.Helper()
	_, err := api.db.Exec(`
		INSERT INTO classi (id, nome, descrizione, documentazione_di_riferimento, dado_vita)
		VALUES ($1, $2, $3, $4, $5)
	`, id, nome, "Test description for "+nome, testDocRiferimento, string(dadoVita))
	if err != nil {
		t.Fatalf("failed to insert classe: %v", err)
	}
}

func (api *testAPI) insertSottoclasse(t *testing.T, id, nome, classeID string) {
	t.Helper()
	_, err := api.db.Exec(`
		INSERT INTO sottoclassi (id, nome, descrizione, documentazione_di_riferimento, id_classe_associata)
		VALUES ($1, $2, $3, $4, $5)
	`, id, nome, "Test description for "+nome, testDocRiferimento, classeID)
	if err != nil {
		t.Fatalf("failed to insert sottoclasse: %v", err)
	}
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
	api := setupTestAPI(t)
	defer api.cleanup(t)

	resp, err := http.Get(api.server.URL + "/health")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result["status"])
	}
}

func TestE2E_ListClassi(t *testing.T) {
	api := setupTestAPI(t)
	defer api.cleanup(t)

	t.Run("empty list", func(t *testing.T) {
		api.truncateTables(t)

		resp, err := http.Get(api.server.URL + "/v1/classi")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result classi.ListClassiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.NumeroDiElementi != 0 {
			t.Errorf("expected 0 elements, got %d", result.NumeroDiElementi)
		}
	})

	t.Run("with data", func(t *testing.T) {
		api.truncateTables(t)
		api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
		api.insertClasse(t, "mago", "Mago", classi.D6)
		api.insertClasse(t, "guerriero", "Guerriero", classi.D10)

		resp, err := http.Get(api.server.URL + "/v1/classi")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result classi.ListClassiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.NumeroDiElementi != 3 {
			t.Errorf("expected 3 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Classi) != 3 {
			t.Errorf("expected 3 classi, got %d", len(result.Classi))
		}
	})

	t.Run("with name filter", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi?nome=bar")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result classi.ListClassiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.NumeroDiElementi != 1 {
			t.Errorf("expected 1 element, got %d", result.NumeroDiElementi)
		}
		if result.Classi[0].ID != "barbaro" {
			t.Errorf("expected 'barbaro', got '%s'", result.Classi[0].ID)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi?$limit=2&$offset=0")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result classi.ListClassiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.NumeroDiElementi != 3 {
			t.Errorf("expected 3 total elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Classi) != 2 {
			t.Errorf("expected 2 classi in page, got %d", len(result.Classi))
		}
	})

	t.Run("with desc sort", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi?sort=desc")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result classi.ListClassiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Classi[0].Nome != "Mago" {
			t.Errorf("expected first classe 'Mago' (desc order), got '%s'", result.Classi[0].Nome)
		}
	})
}

func TestE2E_GetClasse(t *testing.T) {
	api := setupTestAPI(t)
	defer api.cleanup(t)

	api.truncateTables(t)
	api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
	api.insertSottoclasse(t, "berserker", "Berserker", "barbaro")
	api.insertSottoclasse(t, "totemico", "Totemico", "barbaro")

	t.Run("existing classe", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi/barbaro")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result classi.Classe
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

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
		resp, err := http.Get(api.server.URL + "/v1/classi/nonexistent")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}

		var result shared.ErrorObject
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(result.Errors) == 0 {
			t.Fatal("expected error response")
		}
		if result.Errors[0].Code != "NOT_FOUND" {
			t.Errorf("expected code 'NOT_FOUND', got '%s'", result.Errors[0].Code)
		}
	})
}

func TestE2E_ListSottoclassi(t *testing.T) {
	api := setupTestAPI(t)
	defer api.cleanup(t)

	api.truncateTables(t)
	api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
	api.insertSottoclasse(t, "berserker", "Berserker", "barbaro")
	api.insertSottoclasse(t, "totemico", "Totemico", "barbaro")

	t.Run("list sottoclassi", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi/barbaro/sotto-classi")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result classi.ListSottoclassiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.NumeroDiElementi != 2 {
			t.Errorf("expected 2 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Sottoclassi) != 2 {
			t.Errorf("expected 2 sottoclassi, got %d", len(result.Sottoclassi))
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi/nonexistent/sotto-classi")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("with name filter", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi/barbaro/sotto-classi?nome=ber")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result classi.ListSottoclassiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.NumeroDiElementi != 1 {
			t.Errorf("expected 1 element, got %d", result.NumeroDiElementi)
		}
		if result.Sottoclassi[0].ID != "berserker" {
			t.Errorf("expected 'berserker', got '%s'", result.Sottoclassi[0].ID)
		}
	})
}

func TestE2E_GetSottoclasse(t *testing.T) {
	api := setupTestAPI(t)
	defer api.cleanup(t)

	api.truncateTables(t)
	api.insertClasse(t, "barbaro", "Barbaro", classi.D12)
	api.insertSottoclasse(t, "berserker", "Berserker", "barbaro")

	t.Run("existing sottoclasse", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi/barbaro/sotto-classi/berserker")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result classi.SottoClasse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

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
		resp, err := http.Get(api.server.URL + "/v1/classi/nonexistent/sotto-classi/berserker")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("sottoclasse not found", func(t *testing.T) {
		resp, err := http.Get(api.server.URL + "/v1/classi/barbaro/sotto-classi/nonexistent")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
}
