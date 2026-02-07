package transports

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_ListClassi_InvalidFilter(t *testing.T) {
	handler := NewHandler(&mockService{})
	r := chi.NewRouter()
	r.Mount("/classi", handler.Routes())

	t.Run("invalid limit returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/classi?$limit=abc", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("negative offset returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/classi?$offset=-1", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandler_GetClasse_InvalidID(t *testing.T) {
	handler := NewHandler(&mockService{})
	r := chi.NewRouter()
	r.Mount("/classi", handler.Routes())

	t.Run("special characters return 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/classi/inv@lid!", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandler_ListSottoclassi_InvalidInputs(t *testing.T) {
	handler := NewHandler(&mockService{})
	r := chi.NewRouter()
	r.Mount("/classi", handler.Routes())

	t.Run("invalid classe ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/classi/inv@lid/sotto-classi", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("invalid filter returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro/sotto-classi?$limit=abc", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandler_GetSottoclasse_InvalidIDs(t *testing.T) {
	handler := NewHandler(&mockService{})
	r := chi.NewRouter()
	r.Mount("/classi", handler.Routes())

	t.Run("invalid classe ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/classi/inv@lid/sotto-classi/berserker", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("invalid sottoclasse ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/classi/barbaro/sotto-classi/inv@lid!", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
