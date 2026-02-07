package transports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

var benchIncantesimi = []incantesimi.Incantesimo{
	{ID: "dardo-incantato_dnd-2024", Nome: "Dardo Incantato", Livello: 1, ScuolaDiMagia: incantesimi.Evocazione},
	{ID: "palla-di-fuoco_dnd-2024", Nome: "Palla di Fuoco", Livello: 3, ScuolaDiMagia: incantesimi.Evocazione},
	{ID: "individuazione-del-magico_dnd-2024", Nome: "Individuazione del Magico", Livello: 1, ScuolaDiMagia: incantesimi.Divinazione},
	{ID: "luce-danzante_dnd-2024", Nome: "Luce Danzante", Livello: 0, ScuolaDiMagia: incantesimi.Illusione},
	{ID: "scudo_dnd-2024", Nome: "Scudo", Livello: 1, ScuolaDiMagia: incantesimi.Abiurazione},
	{ID: "guarigione_dnd-2024", Nome: "Guarigione", Livello: 1, ScuolaDiMagia: incantesimi.Evocazione},
	{ID: "mano-magica_dnd-2024", Nome: "Mano Magica", Livello: 0, ScuolaDiMagia: incantesimi.Invocazione},
	{ID: "fulmine_dnd-2024", Nome: "Fulmine", Livello: 3, ScuolaDiMagia: incantesimi.Evocazione},
	{ID: "immagine-silenziosa_dnd-2024", Nome: "Immagine Silenziosa", Livello: 1, ScuolaDiMagia: incantesimi.Illusione},
	{ID: "tocco-gelido_dnd-2024", Nome: "Tocco Gelido", Livello: 0, ScuolaDiMagia: incantesimi.Necromamzia},
}

var benchListResponse = &incantesimi.ListIncantesimiResponse{
	PaginationMeta: shared.PaginationMeta{Pagina: 1, NumeroDiElementi: len(benchIncantesimi)},
	Incantesimi:    benchIncantesimi,
}

func setupBenchHandler() http.Handler {
	svc := &mockService{
		listIncantesimiFunc: func(_ context.Context, _ incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error) {
			return benchListResponse, nil
		},
		getIncantesimoFunc: func(_ context.Context, _ string) (*incantesimi.Incantesimo, error) {
			return &benchIncantesimi[0], nil
		},
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	r.Mount("/incantesimi", handler.Routes())
	return r
}

func BenchmarkHandler_ListIncantesimi(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/incantesimi", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHandler_ListIncantesimi_WithParams(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/incantesimi?nome=dardo&$limit=10&$offset=5&sort=desc&livello=1&scuola-di-magia=Evocazione", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHandler_GetIncantesimo(b *testing.B) {
	handler := setupBenchHandler()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/incantesimi/dardo-incantato_dnd-2024", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
