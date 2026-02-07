package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

// migrationsDir returns the absolute path to the migrations directory.
func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}

// setupTestDB starts a Postgres container, runs migrations, and returns
// a connected *sqlx.DB. The container is cleaned up when the test ends.
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(
			filepath.Join(migrationsDir(), "000001_create_classi.up.sql"),
			filepath.Join(migrationsDir(), "000002_create_sottoclassi.up.sql"),
		),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = ctr.Terminate(ctx) })

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return db
}

// seedClasse inserts a classe row into the database.
func seedClasse(t *testing.T, db *sqlx.DB, c classeRow) {
	t.Helper()

	eqJSON, err := json.Marshal(c.EquipaggiamentoPartenza)
	if err != nil {
		t.Fatalf("failed to marshal equipaggiamento: %v", err)
	}

	propJSON, err := json.Marshal(c.ProprietaDiClasse)
	if err != nil {
		t.Fatalf("failed to marshal proprieta: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO classi (id, nome, descrizione, documentazione_di_riferimento, dado_vita,
		                     equipaggiamento_partenza, proprieta_di_classe)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		c.ID, c.Nome, nullStringToPtr(c.Descrizione), c.DocumentazioneDiRiferimento,
		c.DadoVita, eqJSON, propJSON,
	)
	if err != nil {
		t.Fatalf("failed to seed classe %s: %v", c.ID, err)
	}
}

// seedSottoclasse inserts a sottoclasse row into the database.
func seedSottoclasse(t *testing.T, db *sqlx.DB, s sottoclasseRow) {
	t.Helper()

	propJSON, err := json.Marshal(s.ProprietaDiSottoclasse)
	if err != nil {
		t.Fatalf("failed to marshal proprieta: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO sottoclassi (id, nome, descrizione, documentazione_di_riferimento,
		                         id_classe_associata, proprieta_di_sottoclasse)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		s.ID, s.Nome, nullStringToPtr(s.Descrizione), s.DocumentazioneDiRiferimento,
		s.IDClasseAssociata, propJSON,
	)
	if err != nil {
		t.Fatalf("failed to seed sottoclasse %s: %v", s.ID, err)
	}
}

// nullStringToPtr returns nil for invalid NullStrings, or a pointer to
// the string value otherwise.
func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func TestPostgresRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewPostgresRepository(db)
	ctx := context.Background()

	barbaro := classeRow{
		ID: "barbaro", Nome: "Barbaro",
		DocumentazioneDiRiferimento: "DND 2024",
		DadoVita:                    "d12",
	}
	mago := classeRow{
		ID: "mago", Nome: "Mago",
		DocumentazioneDiRiferimento: "DND 2024",
		DadoVita:                    "d6",
	}
	guerriero := classeRow{
		ID: "guerriero", Nome: "Guerriero",
		DocumentazioneDiRiferimento: "DND 2014",
		DadoVita:                    "d10",
	}

	seedClasse(t, db, barbaro)
	seedClasse(t, db, mago)
	seedClasse(t, db, guerriero)

	seedSottoclasse(t, db, sottoclasseRow{
		ID: "berserker", Nome: "Berserker",
		DocumentazioneDiRiferimento: "DND 2024",
		IDClasseAssociata:           "barbaro",
	})

	t.Run("returns all classi with default filter", func(t *testing.T) {
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(result) != 3 {
			t.Errorf("expected 3 classi, got %d", len(result))
		}
		// Sorted by nome ASC: Barbaro, Guerriero, Mago
		if result[0].Nome != "Barbaro" {
			t.Errorf("expected first classe 'Barbaro', got %q", result[0].Nome)
		}
		if result[2].Nome != "Mago" {
			t.Errorf("expected last classe 'Mago', got %q", result[2].Nome)
		}
	})

	t.Run("filters by nome", func(t *testing.T) {
		nome := "mag"
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc, Nome: &nome}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 classe, got %d", len(result))
		}
		if result[0].ID != "mago" {
			t.Errorf("expected classe 'mago', got %q", result[0].ID)
		}
	})

	t.Run("filters by documentazione di riferimento", func(t *testing.T) {
		filter := shared.ListFilter{
			Limit: 20, Offset: 0, Sort: shared.SortAsc,
			DocumentazioneDiRiferimento: []string{"DND 2014"},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 classe, got %d", len(result))
		}
		if result[0].ID != "guerriero" {
			t.Errorf("expected classe 'guerriero', got %q", result[0].ID)
		}
	})

	t.Run("pagination works", func(t *testing.T) {
		filter := shared.ListFilter{Limit: 2, Offset: 0, Sort: shared.SortAsc}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 classi (page 1), got %d", len(result))
		}

		// Second page
		filter.Offset = 2
		result, total, err = repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 classe (page 2), got %d", len(result))
		}
	})

	t.Run("sort desc", func(t *testing.T) {
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortDesc}

		result, _, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Sorted by nome DESC: Mago, Guerriero, Barbaro
		if result[0].Nome != "Mago" {
			t.Errorf("expected first classe 'Mago', got %q", result[0].Nome)
		}
		if result[2].Nome != "Barbaro" {
			t.Errorf("expected last classe 'Barbaro', got %q", result[2].Nome)
		}
	})

	t.Run("includes sottoclassi riferimenti", func(t *testing.T) {
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc}

		result, _, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var barbaroResult *classi.Classe
		for i := range result {
			if result[i].ID == "barbaro" {
				barbaroResult = &result[i]
				break
			}
		}
		if barbaroResult == nil {
			t.Fatal("expected to find barbaro in results")
		}
		if len(barbaroResult.ElencoSottoclassi) != 1 {
			t.Errorf("expected 1 sottoclasse riferimento, got %d", len(barbaroResult.ElencoSottoclassi))
		}
		if barbaroResult.ElencoSottoclassi[0].IDSottoclasse != "berserker" {
			t.Errorf("expected sottoclasse 'berserker', got %q", barbaroResult.ElencoSottoclassi[0].IDSottoclasse)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		nome := "zzz-nonexistent"
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc, Nome: &nome}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 classi, got %d", len(result))
		}
	})
}

func TestPostgresRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewPostgresRepository(db)
	ctx := context.Background()

	equip := equipaggiamentoPartenzaJSON{
		OpzioneA: []classi.OggettoPartenza{
			{Nome: "Ascia Bipenne", Quantita: 1},
		},
		OpzioneB: &classi.Importo{Quantita: 10, Valuta: classi.MO},
	}

	proprieta := proprietaLivelloSlice{
		{LivelloClasse: 1, TrattoDiClasse: &classi.Tratto{Nome: "Ira", Descrizione: "Entra in ira"}},
	}

	seedClasse(t, db, classeRow{
		ID: "barbaro", Nome: "Barbaro",
		DocumentazioneDiRiferimento: "DND 2024",
		DadoVita:                    "d12",
		EquipaggiamentoPartenza:     equip,
		ProprietaDiClasse:           proprieta,
	})

	seedSottoclasse(t, db, sottoclasseRow{
		ID: "berserker", Nome: "Berserker",
		DocumentazioneDiRiferimento: "DND 2024",
		IDClasseAssociata:           "barbaro",
	})

	t.Run("found", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "barbaro")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.ID != "barbaro" {
			t.Errorf("expected id 'barbaro', got %q", result.ID)
		}
		if result.Nome != "Barbaro" {
			t.Errorf("expected nome 'Barbaro', got %q", result.Nome)
		}
		if result.DadoVita != classi.D12 {
			t.Errorf("expected dado vita 'd12', got %q", result.DadoVita)
		}
	})

	t.Run("JSONB equipaggiamento decoded correctly", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "barbaro")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.EquipaggiamentoPartenza == nil {
			t.Fatal("expected non-nil equipaggiamento")
		}
		if len(result.EquipaggiamentoPartenza.OpzioneA) != 1 {
			t.Fatalf("expected 1 oggetto in opzione A, got %d", len(result.EquipaggiamentoPartenza.OpzioneA))
		}
		if result.EquipaggiamentoPartenza.OpzioneA[0].Nome != "Ascia Bipenne" {
			t.Errorf("expected oggetto 'Ascia Bipenne', got %q", result.EquipaggiamentoPartenza.OpzioneA[0].Nome)
		}
		if result.EquipaggiamentoPartenza.OpzioneB == nil {
			t.Fatal("expected non-nil opzione B")
		}
		if result.EquipaggiamentoPartenza.OpzioneB.Quantita != 10 {
			t.Errorf("expected quantita 10, got %d", result.EquipaggiamentoPartenza.OpzioneB.Quantita)
		}
	})

	t.Run("JSONB proprieta decoded correctly", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "barbaro")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.ProprietaDiClasse) != 1 {
			t.Fatalf("expected 1 proprieta, got %d", len(result.ProprietaDiClasse))
		}
		if result.ProprietaDiClasse[0].LivelloClasse != 1 {
			t.Errorf("expected livello 1, got %d", result.ProprietaDiClasse[0].LivelloClasse)
		}
		if result.ProprietaDiClasse[0].TrattoDiClasse == nil {
			t.Fatal("expected non-nil tratto")
		}
		if result.ProprietaDiClasse[0].TrattoDiClasse.Nome != "Ira" {
			t.Errorf("expected tratto 'Ira', got %q", result.ProprietaDiClasse[0].TrattoDiClasse.Nome)
		}
	})

	t.Run("includes sottoclassi riferimenti", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "barbaro")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.ElencoSottoclassi) != 1 {
			t.Fatalf("expected 1 sottoclasse, got %d", len(result.ElencoSottoclassi))
		}
		if result.ElencoSottoclassi[0].IDSottoclasse != "berserker" {
			t.Errorf("expected sottoclasse 'berserker', got %q", result.ElencoSottoclassi[0].IDSottoclasse)
		}
	})

	t.Run("not found returns nil", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "nonexistent")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %+v", result)
		}
	})

	t.Run("nil JSONB fields handled", func(t *testing.T) {
		seedClasse(t, db, classeRow{
			ID: "ladro", Nome: "Ladro",
			DocumentazioneDiRiferimento: "DND 2024",
			DadoVita:                    "d8",
		})

		result, err := repo.GetByID(ctx, "ladro")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.EquipaggiamentoPartenza != nil {
			t.Errorf("expected nil equipaggiamento, got %+v", result.EquipaggiamentoPartenza)
		}
		if result.ProprietaDiClasse != nil {
			t.Errorf("expected nil proprieta, got %+v", result.ProprietaDiClasse)
		}
	})
}

func TestPostgresRepository_ListSottoclassi(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewPostgresRepository(db)
	ctx := context.Background()

	seedClasse(t, db, classeRow{
		ID: "barbaro", Nome: "Barbaro",
		DocumentazioneDiRiferimento: "DND 2024",
		DadoVita:                    "d12",
	})

	seedSottoclasse(t, db, sottoclasseRow{
		ID: "berserker", Nome: "Berserker",
		DocumentazioneDiRiferimento: "DND 2024",
		IDClasseAssociata:           "barbaro",
	})
	seedSottoclasse(t, db, sottoclasseRow{
		ID: "totemico", Nome: "Totemico",
		DocumentazioneDiRiferimento: "DND 2024",
		IDClasseAssociata:           "barbaro",
	})

	t.Run("returns sottoclassi for classe", func(t *testing.T) {
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc}

		result, total, err := repo.ListSottoclassi(ctx, "barbaro", filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 sottoclassi, got %d", len(result))
		}
		// Sorted by nome ASC: Berserker, Totemico
		if result[0].Nome != "Berserker" {
			t.Errorf("expected first 'Berserker', got %q", result[0].Nome)
		}
		if result[1].Nome != "Totemico" {
			t.Errorf("expected second 'Totemico', got %q", result[1].Nome)
		}
	})

	t.Run("returns empty for unknown classe", func(t *testing.T) {
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc}

		result, total, err := repo.ListSottoclassi(ctx, "nonexistent", filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 sottoclassi, got %d", len(result))
		}
	})

	t.Run("filters by nome", func(t *testing.T) {
		nome := "bers"
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc, Nome: &nome}

		result, total, err := repo.ListSottoclassi(ctx, "barbaro", filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 sottoclasse, got %d", len(result))
		}
		if result[0].ID != "berserker" {
			t.Errorf("expected 'berserker', got %q", result[0].ID)
		}
	})

	t.Run("pagination works", func(t *testing.T) {
		filter := shared.ListFilter{Limit: 1, Offset: 0, Sort: shared.SortAsc}

		result, total, err := repo.ListSottoclassi(ctx, "barbaro", filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 sottoclasse (page 1), got %d", len(result))
		}
	})
}

func TestPostgresRepository_GetSottoclasseByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewPostgresRepository(db)
	ctx := context.Background()

	seedClasse(t, db, classeRow{
		ID: "barbaro", Nome: "Barbaro",
		DocumentazioneDiRiferimento: "DND 2024",
		DadoVita:                    "d12",
	})

	proprieta := proprietaLivelloSlice{
		{LivelloClasse: 3, TrattoDiClasse: &classi.Tratto{Nome: "Frenesia", Descrizione: "Attacca con frenesia"}},
	}

	seedSottoclasse(t, db, sottoclasseRow{
		ID: "berserker", Nome: "Berserker",
		DocumentazioneDiRiferimento: "DND 2024",
		IDClasseAssociata:           "barbaro",
		ProprietaDiSottoclasse:      proprieta,
	})

	t.Run("found", func(t *testing.T) {
		result, err := repo.GetSottoclasseByID(ctx, "barbaro", "berserker")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.ID != "berserker" {
			t.Errorf("expected id 'berserker', got %q", result.ID)
		}
		if result.Nome != "Berserker" {
			t.Errorf("expected nome 'Berserker', got %q", result.Nome)
		}
		if result.IDClasseAssociata != "barbaro" {
			t.Errorf("expected classe associata 'barbaro', got %q", result.IDClasseAssociata)
		}
	})

	t.Run("JSONB proprieta decoded correctly", func(t *testing.T) {
		result, err := repo.GetSottoclasseByID(ctx, "barbaro", "berserker")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.ProprietaDiSottoclasse) != 1 {
			t.Fatalf("expected 1 proprieta, got %d", len(result.ProprietaDiSottoclasse))
		}
		if result.ProprietaDiSottoclasse[0].TrattoDiClasse.Nome != "Frenesia" {
			t.Errorf("expected tratto 'Frenesia', got %q", result.ProprietaDiSottoclasse[0].TrattoDiClasse.Nome)
		}
	})

	t.Run("not found returns nil", func(t *testing.T) {
		result, err := repo.GetSottoclasseByID(ctx, "barbaro", "nonexistent")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %+v", result)
		}
	})

	t.Run("wrong classe returns nil", func(t *testing.T) {
		result, err := repo.GetSottoclasseByID(ctx, "wrong-classe", "berserker")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result for wrong classe, got %+v", result)
		}
	})
}

func TestScanJSON(t *testing.T) {
	t.Run("nil source returns nil", func(t *testing.T) {
		var dest []string
		if err := scanJSON(nil, &dest); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != nil {
			t.Errorf("expected nil dest, got %v", dest)
		}
	})

	t.Run("valid JSON bytes", func(t *testing.T) {
		var dest []string
		src := []byte(`["a","b"]`)
		if err := scanJSON(src, &dest); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dest) != 2 || dest[0] != "a" || dest[1] != "b" {
			t.Errorf("unexpected result: %v", dest)
		}
	})

	t.Run("non-byte source returns error", func(t *testing.T) {
		var dest []string
		if err := scanJSON("not bytes", &dest); err == nil {
			t.Fatal("expected error for non-byte source")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		var dest []string
		src := []byte(`{invalid}`)
		if err := scanJSON(src, &dest); err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestProprietaLivelloSlice_ScanValue(t *testing.T) {
	t.Run("scan nil sets nil", func(t *testing.T) {
		var p proprietaLivelloSlice
		if err := p.Scan(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != nil {
			t.Errorf("expected nil, got %v", p)
		}
	})

	t.Run("scan valid JSON", func(t *testing.T) {
		var p proprietaLivelloSlice
		src := []byte(`[{"livello-classe":1}]`)
		if err := p.Scan(src); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(p) != 1 || p[0].LivelloClasse != 1 {
			t.Errorf("unexpected result: %v", p)
		}
	})

	t.Run("value nil returns nil", func(t *testing.T) {
		var p proprietaLivelloSlice
		val, err := p.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil value, got %v", val)
		}
	})

	t.Run("value non-nil returns JSON", func(t *testing.T) {
		p := proprietaLivelloSlice{{LivelloClasse: 1}}
		val, err := p.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == nil {
			t.Fatal("expected non-nil value")
		}
	})
}

func TestNewPaginatedQuery(t *testing.T) {
	t.Run("basic query without filters", func(t *testing.T) {
		args := make(map[string]any)
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc}

		q := newPaginatedQuery("SELECT * FROM t WHERE 1=1", "SELECT COUNT(*) FROM t WHERE 1=1", args, filter)

		if q.args["limit"] != 20 {
			t.Errorf("expected limit 20, got %v", q.args["limit"])
		}
		if q.args["offset"] != 0 {
			t.Errorf("expected offset 0, got %v", q.args["offset"])
		}
	})

	t.Run("with nome filter", func(t *testing.T) {
		args := make(map[string]any)
		nome := "test%val"
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc, Nome: &nome}

		q := newPaginatedQuery("SELECT * FROM t WHERE 1=1", "SELECT COUNT(*) FROM t WHERE 1=1", args, filter)

		nomeArg, ok := q.args["nome"].(string)
		if !ok {
			t.Fatal("expected nome arg to be string")
		}
		// % in the user input should be escaped
		if nomeArg != `%test\%val%` {
			t.Errorf("expected escaped nome arg, got %q", nomeArg)
		}
	})

	t.Run("desc sort", func(t *testing.T) {
		args := make(map[string]any)
		filter := shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortDesc}

		q := newPaginatedQuery("SELECT * FROM t WHERE 1=1", "SELECT COUNT(*) FROM t WHERE 1=1", args, filter)

		if q.query == "" {
			t.Fatal("expected non-empty query")
		}
		// The query should contain DESC
		if !contains(q.query, "DESC") {
			t.Errorf("expected query to contain DESC, got %q", q.query)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
