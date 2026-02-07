package persistence

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}

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
			filepath.Join(migrationsDir(), "000003_create_incantesimi.up.sql"),
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

type seedIncantesimo struct {
	ID                          string
	Nome                        string
	Livello                     int32
	ScuolaDiMagia               string
	TempoDiLancio               string
	Gittata                     string
	Area                        string
	Concentrazione              bool
	SemprePreparato             bool
	Rituale                     bool
	Componenti                  []string
	ComponentiMateriali         string
	Durata                      string
	Descrizione                 string
	EffettoIncantesimo          json.RawMessage
	EffettoLivelloMaggiore      json.RawMessage
	Classi                      string
	DocumentazioneDiRiferimento string
}

func seedIncantesimoRow(t *testing.T, db *sqlx.DB, s seedIncantesimo) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO incantesimi (id, nome, livello, scuola_di_magia, tempo_di_lancio, gittata, area,
		                          concentrazione, sempre_preparato, rituale, componenti, componenti_materiali,
		                          durata, descrizione, effetto_incantesimo, effetto_livello_maggiore,
		                          classi, documentazione_di_riferimento)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		s.ID, s.Nome, s.Livello, s.ScuolaDiMagia, s.TempoDiLancio, s.Gittata,
		nilIfEmpty(s.Area), s.Concentrazione, s.SemprePreparato, s.Rituale,
		pq.Array(s.Componenti), nilIfEmpty(s.ComponentiMateriali),
		s.Durata, s.Descrizione, nilRawJSON(s.EffettoIncantesimo), nilRawJSON(s.EffettoLivelloMaggiore),
		s.Classi, s.DocumentazioneDiRiferimento,
	)
	if err != nil {
		t.Fatalf("failed to seed incantesimo %s: %v", s.ID, err)
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nilRawJSON(data json.RawMessage) *string {
	if len(data) == 0 {
		return nil
	}
	s := string(data)
	return &s
}

func TestPostgresRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewPostgresRepository(db)
	ctx := context.Background()

	dardoIncantato := seedIncantesimo{
		ID: "dardo-incantato_dnd-2024", Nome: "Dardo Incantato", Livello: 1,
		ScuolaDiMagia: "Evocazione", TempoDiLancio: "1 azione",
		Gittata: "36 metri", Concentrazione: false, Rituale: false,
		Componenti: []string{"V", "S"}, Durata: "Istantanea",
		Descrizione: "Crei tre dardi luminosi di forza magica.",
		Classi:      "Mago", DocumentazioneDiRiferimento: "DND 2024",
	}

	pallaDiFuoco := seedIncantesimo{
		ID: "palla-di-fuoco_dnd-2024", Nome: "Palla di Fuoco", Livello: 3,
		ScuolaDiMagia: "Evocazione", TempoDiLancio: "1 azione",
		Gittata: "45 metri", Area: "Sfera di 6 metri di raggio",
		Concentrazione: false, Rituale: false,
		Componenti: []string{"V", "S", "M"}, ComponentiMateriali: "Una piccola palla di sterco di pipistrello e zolfo",
		Durata: "Istantanea", Descrizione: "Un lampo di luce si emana dal tuo dito.",
		Classi: "Mago, Stregone", DocumentazioneDiRiferimento: "DND 2024",
	}

	individuazioneMagia := seedIncantesimo{
		ID: "individuazione-del-magico_dnd-2024", Nome: "Individuazione del Magico", Livello: 1,
		ScuolaDiMagia: "Divinazione", TempoDiLancio: "1 azione, Rituale",
		Gittata: "Incantatore", Concentrazione: true, Rituale: true,
		Componenti: []string{"V", "S"}, Durata: "Concentrazione, fino a 10 minuti",
		Descrizione:                 "Percepisci la presenza di magia entro 9 metri.",
		Classi:                      "Bardo, Chierico, Druido, Mago, Paladino, Ranger, Stregone",
		DocumentazioneDiRiferimento: "DND 2024",
	}

	luceDanzante := seedIncantesimo{
		ID: "luce-danzante_dnd-2024", Nome: "Luce Danzante", Livello: 0,
		ScuolaDiMagia: "Illusione", TempoDiLancio: "1 azione",
		Gittata: "36 metri", Concentrazione: true, SemprePreparato: true,
		Rituale: false, Componenti: []string{"V", "S", "M"},
		ComponentiMateriali: "Un po' di fosforo o un pezzo di legno d'olmo",
		Durata:              "Concentrazione, fino a 1 minuto",
		Descrizione:         "Crei fino a quattro luci delle dimensioni di una torcia.",
		Classi:              "Bardo, Mago", DocumentazioneDiRiferimento: "DND 2024",
	}

	seedIncantesimoRow(t, db, dardoIncantato)
	seedIncantesimoRow(t, db, pallaDiFuoco)
	seedIncantesimoRow(t, db, individuazioneMagia)
	seedIncantesimoRow(t, db, luceDanzante)

	t.Run("returns all incantesimi with default filter", func(t *testing.T) {
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 4 {
			t.Errorf("expected total 4, got %d", total)
		}
		if len(result) != 4 {
			t.Errorf("expected 4 incantesimi, got %d", len(result))
		}
		// Sorted by nome ASC
		if result[0].Nome != "Dardo Incantato" {
			t.Errorf("expected first 'Dardo Incantato', got %q", result[0].Nome)
		}
	})

	t.Run("filters by nome", func(t *testing.T) {
		nome := "palla"
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc, Nome: &nome},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 incantesimo, got %d", len(result))
		}
		if result[0].ID != "palla-di-fuoco_dnd-2024" {
			t.Errorf("expected 'palla-di-fuoco_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by livello", func(t *testing.T) {
		livello := int32(3)
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Livello:    &livello,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 incantesimo, got %d", len(result))
		}
		if result[0].ID != "palla-di-fuoco_dnd-2024" {
			t.Errorf("expected 'palla-di-fuoco_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by livello 0 (cantrips)", func(t *testing.T) {
		livello := int32(0)
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Livello:    &livello,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if result[0].ID != "luce-danzante_dnd-2024" {
			t.Errorf("expected 'luce-danzante_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by scuola di magia", func(t *testing.T) {
		scuola := "Divinazione"
		filter := incantesimi.IncantesimiFilter{
			ListFilter:    shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			ScuolaDiMagia: &scuola,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if result[0].ID != "individuazione-del-magico_dnd-2024" {
			t.Errorf("expected 'individuazione-del-magico_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by concentrazione", func(t *testing.T) {
		conc := true
		filter := incantesimi.IncantesimiFilter{
			ListFilter:     shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Concentrazione: &conc,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		for _, r := range result {
			if !r.Concentrazione {
				t.Errorf("expected concentrazione=true for %s", r.ID)
			}
		}
	})

	t.Run("filters by rituale", func(t *testing.T) {
		rit := true
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Rituale:    &rit,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if result[0].ID != "individuazione-del-magico_dnd-2024" {
			t.Errorf("expected 'individuazione-del-magico_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by componenti (array containment)", func(t *testing.T) {
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Componenti: []string{"M"},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		for _, r := range result {
			hasM := false
			for _, c := range r.Componenti {
				if c == "M" {
					hasM = true
					break
				}
			}
			if !hasM {
				t.Errorf("expected componenti to contain M for %s", r.ID)
			}
		}
	})

	t.Run("filters by componenti materiali", func(t *testing.T) {
		cm := "fosforo"
		filter := incantesimi.IncantesimiFilter{
			ListFilter:          shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			ComponentiMateriali: &cm,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if result[0].ID != "luce-danzante_dnd-2024" {
			t.Errorf("expected 'luce-danzante_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by tempo di lancio", func(t *testing.T) {
		tdl := "Rituale"
		filter := incantesimi.IncantesimiFilter{
			ListFilter:    shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			TempoDiLancio: &tdl,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if result[0].ID != "individuazione-del-magico_dnd-2024" {
			t.Errorf("expected 'individuazione-del-magico_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by gittata", func(t *testing.T) {
		gittata := "45"
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Gittata:    &gittata,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if result[0].ID != "palla-di-fuoco_dnd-2024" {
			t.Errorf("expected 'palla-di-fuoco_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by durata", func(t *testing.T) {
		durata := "1 minuto"
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Durata:     &durata,
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if result[0].ID != "luce-danzante_dnd-2024" {
			t.Errorf("expected 'luce-danzante_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("filters by classi (single)", func(t *testing.T) {
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Classi:     []string{"Stregone"},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Palla di Fuoco (Mago, Stregone) and Individuazione del Magico (..., Stregone)
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		for _, r := range result {
			if r.ID != "individuazione-del-magico_dnd-2024" && r.ID != "palla-di-fuoco_dnd-2024" {
				t.Errorf("unexpected incantesimo %q in results", r.ID)
			}
		}
	})

	t.Run("filters by classi (multiple AND logic)", func(t *testing.T) {
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc},
			Classi:     []string{"Bardo", "Druido"},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Only Individuazione del Magico has both Bardo AND Druido
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 incantesimo, got %d", len(result))
		}
		if result[0].ID != "individuazione-del-magico_dnd-2024" {
			t.Errorf("expected 'individuazione-del-magico_dnd-2024', got %q", result[0].ID)
		}
	})

	t.Run("pagination works", func(t *testing.T) {
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 2, Offset: 0, Sort: shared.SortAsc},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 4 {
			t.Errorf("expected total 4, got %d", total)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 incantesimi (page 1), got %d", len(result))
		}

		// Second page
		filter.Offset = 2
		result, total, err = repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 4 {
			t.Errorf("expected total 4, got %d", total)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 incantesimi (page 2), got %d", len(result))
		}
	})

	t.Run("sort desc", func(t *testing.T) {
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortDesc},
		}

		result, _, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[0].Nome != "Palla di Fuoco" {
			t.Errorf("expected first 'Palla di Fuoco', got %q", result[0].Nome)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		nome := "zzz-nonexistent"
		filter := incantesimi.IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0, Sort: shared.SortAsc, Nome: &nome},
		}

		result, total, err := repo.List(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 incantesimi, got %d", len(result))
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

	effetto := json.RawMessage(`{"ripetizione-effetto":3,"effetto":[{"tipo_di_danno":["Forza"],"numero_di_dadi":1,"tipo_di_dado":"d4"}]}`)

	seedIncantesimoRow(t, db, seedIncantesimo{
		ID: "dardo-incantato_dnd-2024", Nome: "Dardo Incantato", Livello: 1,
		ScuolaDiMagia: "Evocazione", TempoDiLancio: "1 azione",
		Gittata: "36 metri", Concentrazione: false, Rituale: false,
		Componenti: []string{"V", "S"}, Durata: "Istantanea",
		Descrizione:        "Crei tre dardi luminosi di forza magica.",
		EffettoIncantesimo: effetto,
		Classi:             "Mago", DocumentazioneDiRiferimento: "DND 2024",
	})

	t.Run("found", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "dardo-incantato_dnd-2024")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.ID != "dardo-incantato_dnd-2024" {
			t.Errorf("expected id 'dardo-incantato_dnd-2024', got %q", result.ID)
		}
		if result.Nome != "Dardo Incantato" {
			t.Errorf("expected nome 'Dardo Incantato', got %q", result.Nome)
		}
		if result.Livello != 1 {
			t.Errorf("expected livello 1, got %d", result.Livello)
		}
		if result.ScuolaDiMagia != "Evocazione" {
			t.Errorf("expected scuola 'Evocazione', got %q", result.ScuolaDiMagia)
		}
		if len(result.Componenti) != 2 {
			t.Fatalf("expected 2 componenti, got %d", len(result.Componenti))
		}
		if result.Componenti[0] != "V" || result.Componenti[1] != "S" {
			t.Errorf("expected componenti [V, S], got %v", result.Componenti)
		}
	})

	t.Run("JSONB effetto decoded correctly", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "dardo-incantato_dnd-2024")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.EffettoIncantesimo == nil {
			t.Fatal("expected non-nil effetto incantesimo")
		}
		if result.EffettoIncantesimo.RipetizioneEffetto == nil || *result.EffettoIncantesimo.RipetizioneEffetto != 3 {
			t.Errorf("expected ripetizione effetto 3, got %v", result.EffettoIncantesimo.RipetizioneEffetto)
		}
		if result.EffettoIncantesimo.Effetto == nil {
			t.Fatal("expected non-nil effetto")
		}
	})

	t.Run("componenti array decoded correctly", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "dardo-incantato_dnd-2024")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Componenti) != 2 {
			t.Fatalf("expected 2 componenti, got %d", len(result.Componenti))
		}
		if result.Componenti[0] != "V" {
			t.Errorf("expected first componente 'V', got %q", result.Componenti[0])
		}
		if result.Componenti[1] != "S" {
			t.Errorf("expected second componente 'S', got %q", result.Componenti[1])
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
		seedIncantesimoRow(t, db, seedIncantesimo{
			ID: "luce_dnd-2024", Nome: "Luce", Livello: 0,
			ScuolaDiMagia: "Evocazione", TempoDiLancio: "1 azione",
			Gittata: "Contatto", Concentrazione: false, Rituale: false,
			Componenti: []string{"V"}, Durata: "1 ora",
			Descrizione: "Tocchi un oggetto e lo fai brillare.",
			Classi:      "Chierico, Mago", DocumentazioneDiRiferimento: "DND 2024",
		})

		result, err := repo.GetByID(ctx, "luce_dnd-2024")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.EffettoIncantesimo != nil {
			t.Errorf("expected nil effetto incantesimo, got %+v", result.EffettoIncantesimo)
		}
		if result.EffettoLivelloMaggiore != nil {
			t.Errorf("expected nil effetto livello maggiore, got %+v", result.EffettoLivelloMaggiore)
		}
	})
}
