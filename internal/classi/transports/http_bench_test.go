package transports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

var benchClassi = []classi.Classe{
	{ID: "barbaro", Nome: "Barbaro", DadoVita: classi.D12, Descrizione: "Un feroce guerriero"},
	{ID: "bardo", Nome: "Bardo", DadoVita: classi.D8, Descrizione: "Un artista magico"},
	{ID: "chierico", Nome: "Chierico", DadoVita: classi.D8, Descrizione: "Un sacerdote combattente"},
	{ID: "druido", Nome: "Druido", DadoVita: classi.D8, Descrizione: "Un custode della natura"},
	{ID: "guerriero", Nome: "Guerriero", DadoVita: classi.D10, Descrizione: "Un maestro delle armi"},
	{ID: "ladro", Nome: "Ladro", DadoVita: classi.D8, Descrizione: "Un esperto furtivo"},
	{ID: "mago", Nome: "Mago", DadoVita: classi.D6, Descrizione: "Un incantatore arcano"},
	{ID: "monaco", Nome: "Monaco", DadoVita: classi.D8, Descrizione: "Un artista marziale"},
	{ID: "paladino", Nome: "Paladino", DadoVita: classi.D10, Descrizione: "Un cavaliere sacro"},
	{ID: "ranger", Nome: "Ranger", DadoVita: classi.D10, Descrizione: "Un esploratore"},
}

var benchSottoclassi = []classi.SottoClasse{
	{ID: "berserker", Nome: "Berserker", IDClasseAssociata: "barbaro"},
	{ID: "totemico", Nome: "Totemico", IDClasseAssociata: "barbaro"},
}

func setupBenchHandler() http.Handler {
	repo := &mockRepository{
		listFunc: func(_ context.Context, _ shared.ListFilter) ([]classi.Classe, int, error) {
			return benchClassi, len(benchClassi), nil
		},
		getByIDFunc: func(_ context.Context, _ string) (*classi.Classe, error) {
			return &benchClassi[0], nil
		},
		listSottoclassiFunc: func(_ context.Context, _ string, _ shared.ListFilter) ([]classi.SottoClasse, int, error) {
			return benchSottoclassi, len(benchSottoclassi), nil
		},
		getSottoclasseByIDFunc: func(_ context.Context, _, _ string) (*classi.SottoClasse, error) {
			return &benchSottoclassi[0], nil
		},
	}

	handler := setupTestHandler(repo)
	r := chi.NewRouter()
	r.Mount("/classi", handler.Routes())
	return r
}

func BenchmarkHandler_ListClassi(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/classi", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHandler_ListClassi_WithParams(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/classi?nome=bar&$limit=10&$offset=5&sort=desc", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHandler_GetClasse(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHandler_ListSottoclassi(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro/sotto-classi", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHandler_GetSottoclasse(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro/sotto-classi/berserker", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
