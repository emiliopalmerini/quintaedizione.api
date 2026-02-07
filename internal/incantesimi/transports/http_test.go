package transports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type mockService struct {
	listIncantesimiFunc func(ctx context.Context, filter incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error)
	getIncantesimoFunc  func(ctx context.Context, id string) (*incantesimi.Incantesimo, error)
}

func (m *mockService) ListIncantesimi(ctx context.Context, filter incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error) {
	if m.listIncantesimiFunc != nil {
		return m.listIncantesimiFunc(ctx, filter)
	}
	return &incantesimi.ListIncantesimiResponse{}, nil
}

func (m *mockService) GetIncantesimo(ctx context.Context, id string) (*incantesimi.Incantesimo, error) {
	if m.getIncantesimoFunc != nil {
		return m.getIncantesimoFunc(ctx, id)
	}
	return nil, nil
}

func TestHandler_ListIncantesimi(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockService{
			listIncantesimiFunc: func(_ context.Context, _ incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error) {
				return &incantesimi.ListIncantesimiResponse{
					PaginationMeta: shared.PaginationMeta{Pagina: 1, NumeroDiElementi: 2},
					Incantesimi: []incantesimi.Incantesimo{
						{ID: "dardo-incantato_dnd-2024", Nome: "Dardo Incantato", Livello: 1, ScuolaDiMagia: incantesimi.Evocazione},
						{ID: "palla-di-fuoco_dnd-2024", Nome: "Palla di Fuoco", Livello: 3, ScuolaDiMagia: incantesimi.Evocazione},
					},
				}, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response incantesimi.ListIncantesimiResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.NumeroDiElementi != 2 {
			t.Errorf("expected 2 elements, got %d", response.NumeroDiElementi)
		}
		if len(response.Incantesimi) != 2 {
			t.Errorf("expected 2 incantesimi, got %d", len(response.Incantesimi))
		}
	})

	t.Run("with base query params", func(t *testing.T) {
		var capturedFilter incantesimi.IncantesimiFilter

		svc := &mockService{
			listIncantesimiFunc: func(_ context.Context, filter incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error) {
				capturedFilter = filter
				return &incantesimi.ListIncantesimiResponse{}, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi?nome=dardo&$limit=10&$offset=5&sort=desc", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if capturedFilter.Nome == nil || *capturedFilter.Nome != "dardo" {
			t.Errorf("expected nome filter 'dardo', got %v", capturedFilter.Nome)
		}
		if capturedFilter.Limit != 10 {
			t.Errorf("expected limit 10, got %d", capturedFilter.Limit)
		}
		if capturedFilter.Offset != 5 {
			t.Errorf("expected offset 5, got %d", capturedFilter.Offset)
		}
		if capturedFilter.Sort != shared.SortDesc {
			t.Errorf("expected sort desc, got %s", capturedFilter.Sort)
		}
	})

	t.Run("with domain-specific filters", func(t *testing.T) {
		var capturedFilter incantesimi.IncantesimiFilter

		svc := &mockService{
			listIncantesimiFunc: func(_ context.Context, filter incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error) {
				capturedFilter = filter
				return &incantesimi.ListIncantesimiResponse{}, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet,
			"/incantesimi?livello=3&scuola-di-magia=Evocazione&concentrazione=true&rituale=false&componenti=V&componenti=S", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if capturedFilter.Livello == nil || *capturedFilter.Livello != 3 {
			t.Errorf("expected livello 3, got %v", capturedFilter.Livello)
		}
		if capturedFilter.ScuolaDiMagia == nil || *capturedFilter.ScuolaDiMagia != "Evocazione" {
			t.Errorf("expected scuola Evocazione, got %v", capturedFilter.ScuolaDiMagia)
		}
		if capturedFilter.Concentrazione == nil || *capturedFilter.Concentrazione != true {
			t.Errorf("expected concentrazione true, got %v", capturedFilter.Concentrazione)
		}
		if capturedFilter.Rituale == nil || *capturedFilter.Rituale != false {
			t.Errorf("expected rituale false, got %v", capturedFilter.Rituale)
		}
		if len(capturedFilter.Componenti) != 2 || capturedFilter.Componenti[0] != "V" || capturedFilter.Componenti[1] != "S" {
			t.Errorf("expected componenti [V, S], got %v", capturedFilter.Componenti)
		}
	})

	t.Run("invalid livello", func(t *testing.T) {
		svc := &mockService{}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi?livello=10", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("invalid scuola di magia", func(t *testing.T) {
		svc := &mockService{}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi?scuola-di-magia=InvalidSchool", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("invalid componenti", func(t *testing.T) {
		svc := &mockService{}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi?componenti=X", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockService{
			listIncantesimiFunc: func(_ context.Context, _ incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error) {
				return nil, shared.NewInternalError(nil)
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestHandler_GetIncantesimo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockService{
			getIncantesimoFunc: func(_ context.Context, id string) (*incantesimi.Incantesimo, error) {
				if id == "dardo-incantato_dnd-2024" {
					return &incantesimi.Incantesimo{
						ID: "dardo-incantato_dnd-2024", Nome: "Dardo Incantato",
						Livello: 1, ScuolaDiMagia: incantesimi.Evocazione,
					}, nil
				}
				return nil, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi/dardo-incantato_dnd-2024", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response incantesimi.Incantesimo
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.ID != "dardo-incantato_dnd-2024" {
			t.Errorf("expected id 'dardo-incantato_dnd-2024', got '%s'", response.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{
			getIncantesimoFunc: func(_ context.Context, id string) (*incantesimi.Incantesimo, error) {
				return nil, incantesimi.ErrIncantesimoNotFound(id)
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi/nonexistent", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := &mockService{}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/incantesimi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/incantesimi/invalid%20id%21", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
