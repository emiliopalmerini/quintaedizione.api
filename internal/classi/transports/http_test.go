package transports

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type mockRepository struct {
	listFunc               func(ctx context.Context, filter shared.ListFilter) ([]classi.Classe, int, error)
	getByIDFunc            func(ctx context.Context, id string) (*classi.Classe, error)
	listSottoclassiFunc    func(ctx context.Context, classeID string, filter shared.ListFilter) ([]classi.SottoClasse, int, error)
	getSottoclasseByIDFunc func(ctx context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error)
}

func (m *mockRepository) List(ctx context.Context, filter shared.ListFilter) ([]classi.Classe, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*classi.Classe, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) ListSottoclassi(ctx context.Context, classeID string, filter shared.ListFilter) ([]classi.SottoClasse, int, error) {
	if m.listSottoclassiFunc != nil {
		return m.listSottoclassiFunc(ctx, classeID, filter)
	}
	return nil, 0, nil
}

func (m *mockRepository) GetSottoclasseByID(ctx context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error) {
	if m.getSottoclasseByIDFunc != nil {
		return m.getSottoclasseByIDFunc(ctx, classeID, sottoclasseID)
	}
	return nil, nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func setupTestHandler(repo classi.Repository) *Handler {
	logger := newTestLogger()
	service := classi.NewService(repo, logger)
	return NewHandler(service)
}

func TestHandler_ListClassi(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedClassi := []classi.Classe{
			{ID: "barbaro", Nome: "Barbaro", DadoVita: classi.D12},
			{ID: "mago", Nome: "Mago", DadoVita: classi.D6},
		}

		repo := &mockRepository{
			listFunc: func(_ context.Context, _ shared.ListFilter) ([]classi.Classe, int, error) {
				return expectedClassi, 2, nil
			},
		}

		handler := setupTestHandler(repo)
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

		repo := &mockRepository{
			listFunc: func(_ context.Context, filter shared.ListFilter) ([]classi.Classe, int, error) {
				capturedFilter = filter
				return []classi.Classe{}, 0, nil
			},
		}

		handler := setupTestHandler(repo)
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

	t.Run("repository error", func(t *testing.T) {
		repo := &mockRepository{
			listFunc: func(_ context.Context, _ shared.ListFilter) ([]classi.Classe, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		handler := setupTestHandler(repo)
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
		expectedClasse := &classi.Classe{
			ID:       "barbaro",
			Nome:     "Barbaro",
			DadoVita: classi.D12,
		}

		repo := &mockRepository{
			getByIDFunc: func(_ context.Context, id string) (*classi.Classe, error) {
				if id == "barbaro" {
					return expectedClasse, nil
				}
				return nil, nil
			},
		}

		handler := setupTestHandler(repo)
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
		repo := &mockRepository{
			getByIDFunc: func(_ context.Context, _ string) (*classi.Classe, error) {
				return nil, nil
			},
		}

		handler := setupTestHandler(repo)
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
		parentClasse := &classi.Classe{ID: "barbaro", Nome: "Barbaro"}
		expectedSottoclassi := []classi.SottoClasse{
			{ID: "berserker", Nome: "Berserker", IDClasseAssociata: "barbaro"},
		}

		repo := &mockRepository{
			getByIDFunc: func(_ context.Context, id string) (*classi.Classe, error) {
				if id == "barbaro" {
					return parentClasse, nil
				}
				return nil, nil
			},
			listSottoclassiFunc: func(_ context.Context, classeID string, _ shared.ListFilter) ([]classi.SottoClasse, int, error) {
				if classeID == "barbaro" {
					return expectedSottoclassi, 1, nil
				}
				return nil, 0, nil
			},
		}

		handler := setupTestHandler(repo)
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
		repo := &mockRepository{
			getByIDFunc: func(_ context.Context, _ string) (*classi.Classe, error) {
				return nil, nil
			},
		}

		handler := setupTestHandler(repo)
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
		parentClasse := &classi.Classe{ID: "barbaro", Nome: "Barbaro"}
		expectedSottoclasse := &classi.SottoClasse{
			ID:                "berserker",
			Nome:              "Berserker",
			IDClasseAssociata: "barbaro",
		}

		repo := &mockRepository{
			getByIDFunc: func(_ context.Context, id string) (*classi.Classe, error) {
				if id == "barbaro" {
					return parentClasse, nil
				}
				return nil, nil
			},
			getSottoclasseByIDFunc: func(_ context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error) {
				if classeID == "barbaro" && sottoclasseID == "berserker" {
					return expectedSottoclasse, nil
				}
				return nil, nil
			},
		}

		handler := setupTestHandler(repo)
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
		repo := &mockRepository{
			getByIDFunc: func(_ context.Context, _ string) (*classi.Classe, error) {
				return nil, nil
			},
		}

		handler := setupTestHandler(repo)
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
		parentClasse := &classi.Classe{ID: "barbaro", Nome: "Barbaro"}

		repo := &mockRepository{
			getByIDFunc: func(_ context.Context, _ string) (*classi.Classe, error) {
				return parentClasse, nil
			},
			getSottoclasseByIDFunc: func(_ context.Context, _, _ string) (*classi.SottoClasse, error) {
				return nil, nil
			},
		}

		handler := setupTestHandler(repo)
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
