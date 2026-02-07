package transports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type mockService struct {
	listClassiFunc      func(ctx context.Context, filter shared.ListFilter) (*classi.ListClassiResponse, error)
	getClasseFunc       func(ctx context.Context, id string) (*classi.Classe, error)
	listSottoclassiFunc func(ctx context.Context, classeID string, filter shared.ListFilter) (*classi.ListSottoclassiResponse, error)
	getSottoclasseFunc  func(ctx context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error)
}

func (m *mockService) ListClassi(ctx context.Context, filter shared.ListFilter) (*classi.ListClassiResponse, error) {
	if m.listClassiFunc != nil {
		return m.listClassiFunc(ctx, filter)
	}
	return &classi.ListClassiResponse{}, nil
}

func (m *mockService) GetClasse(ctx context.Context, id string) (*classi.Classe, error) {
	if m.getClasseFunc != nil {
		return m.getClasseFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockService) ListSottoclassi(ctx context.Context, classeID string, filter shared.ListFilter) (*classi.ListSottoclassiResponse, error) {
	if m.listSottoclassiFunc != nil {
		return m.listSottoclassiFunc(ctx, classeID, filter)
	}
	return &classi.ListSottoclassiResponse{}, nil
}

func (m *mockService) GetSottoclasse(ctx context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error) {
	if m.getSottoclasseFunc != nil {
		return m.getSottoclasseFunc(ctx, classeID, sottoclasseID)
	}
	return nil, nil
}

func TestHandler_ListClassi(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockService{
			listClassiFunc: func(_ context.Context, _ shared.ListFilter) (*classi.ListClassiResponse, error) {
				return &classi.ListClassiResponse{
					Pagina:           1,
					NumeroDiElementi: 2,
					Classi: []classi.Classe{
						{ID: "barbaro", Nome: "Barbaro", DadoVita: classi.D12},
						{ID: "mago", Nome: "Mago", DadoVita: classi.D6},
					},
				}, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response classi.ListClassiResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.NumeroDiElementi != 2 {
			t.Errorf("expected 2 elements, got %d", response.NumeroDiElementi)
		}
		if len(response.Classi) != 2 {
			t.Errorf("expected 2 classi, got %d", len(response.Classi))
		}
	})

	t.Run("with query params", func(t *testing.T) {
		var capturedFilter shared.ListFilter

		svc := &mockService{
			listClassiFunc: func(_ context.Context, filter shared.ListFilter) (*classi.ListClassiResponse, error) {
				capturedFilter = filter
				return &classi.ListClassiResponse{}, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi?nome=bar&$limit=10&$offset=5&sort=desc", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if capturedFilter.Nome == nil || *capturedFilter.Nome != "bar" {
			t.Errorf("expected nome filter 'bar', got %v", capturedFilter.Nome)
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

	t.Run("service error", func(t *testing.T) {
		svc := &mockService{
			listClassiFunc: func(_ context.Context, _ shared.ListFilter) (*classi.ListClassiResponse, error) {
				return nil, shared.NewInternalError(nil)
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestHandler_GetClasse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockService{
			getClasseFunc: func(_ context.Context, id string) (*classi.Classe, error) {
				if id == "barbaro" {
					return &classi.Classe{ID: "barbaro", Nome: "Barbaro", DadoVita: classi.D12}, nil
				}
				return nil, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response classi.Classe
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.ID != "barbaro" {
			t.Errorf("expected id 'barbaro', got '%s'", response.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockService{
			getClasseFunc: func(_ context.Context, id string) (*classi.Classe, error) {
				return nil, classi.ErrClasseNotFound(id)
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi/nonexistent", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestHandler_ListSottoclassi(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockService{
			listSottoclassiFunc: func(_ context.Context, classeID string, _ shared.ListFilter) (*classi.ListSottoclassiResponse, error) {
				if classeID == "barbaro" {
					return &classi.ListSottoclassiResponse{
						Pagina:           1,
						NumeroDiElementi: 1,
						Sottoclassi: []classi.SottoClasse{
							{ID: "berserker", Nome: "Berserker", IDClasseAssociata: "barbaro"},
						},
					}, nil
				}
				return &classi.ListSottoclassiResponse{}, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro/sotto-classi", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response classi.ListSottoclassiResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.NumeroDiElementi != 1 {
			t.Errorf("expected 1 element, got %d", response.NumeroDiElementi)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		svc := &mockService{
			listSottoclassiFunc: func(_ context.Context, classeID string, _ shared.ListFilter) (*classi.ListSottoclassiResponse, error) {
				return nil, classi.ErrClasseNotFound(classeID)
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi/nonexistent/sotto-classi", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestHandler_GetSottoclasse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockService{
			getSottoclasseFunc: func(_ context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error) {
				if classeID == "barbaro" && sottoclasseID == "berserker" {
					return &classi.SottoClasse{
						ID:                "berserker",
						Nome:              "Berserker",
						IDClasseAssociata: "barbaro",
					}, nil
				}
				return nil, nil
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro/sotto-classi/berserker", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response classi.SottoClasse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.ID != "berserker" {
			t.Errorf("expected id 'berserker', got '%s'", response.ID)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		svc := &mockService{
			getSottoclasseFunc: func(_ context.Context, classeID, _ string) (*classi.SottoClasse, error) {
				return nil, classi.ErrClasseNotFound(classeID)
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi/nonexistent/sotto-classi/berserker", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("sottoclasse not found", func(t *testing.T) {
		svc := &mockService{
			getSottoclasseFunc: func(_ context.Context, _, sottoclasseID string) (*classi.SottoClasse, error) {
				return nil, classi.ErrSottoclasseNotFound(sottoclasseID)
			},
		}

		handler := NewHandler(svc)
		r := chi.NewRouter()
		r.Mount("/classi", handler.Routes())

		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro/sotto-classi/nonexistent", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}
