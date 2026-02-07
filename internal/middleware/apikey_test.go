package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func TestAPIKey(t *testing.T) {
	const validKey = "test-secret-key"

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("valid key passes through", func(t *testing.T) {
		handler := APIKey(validKey)(okHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(APIKeyHeader, validKey)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("missing key returns 401", func(t *testing.T) {
		handler := APIKey(validKey)(okHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}

		var errResp shared.ErrorObject
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if len(errResp.Errors) == 0 {
			t.Fatal("expected at least one error in response")
		}
		if errResp.Errors[0].Detail != "missing API key" {
			t.Errorf("expected detail 'missing API key', got %q", errResp.Errors[0].Detail)
		}
	})

	t.Run("invalid key returns 401", func(t *testing.T) {
		handler := APIKey(validKey)(okHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(APIKeyHeader, "wrong-key")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}

		var errResp shared.ErrorObject
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if len(errResp.Errors) == 0 {
			t.Fatal("expected at least one error in response")
		}
		if errResp.Errors[0].Detail != "invalid API key" {
			t.Errorf("expected detail 'invalid API key', got %q", errResp.Errors[0].Detail)
		}
	})

	t.Run("empty configured key skips auth", func(t *testing.T) {
		handler := APIKey("")(okHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("response has correct content type", func(t *testing.T) {
		handler := APIKey(validKey)(okHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		ct := rec.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", ct)
		}
	})
}
