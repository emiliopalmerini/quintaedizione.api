package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) PingContext(_ context.Context) error {
	return m.err
}

func TestHandler_ServeHTTP(t *testing.T) {
	t.Run("healthy database returns 200", func(t *testing.T) {
		h := NewHandler(&mockPinger{}, "1.0.0")
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp Response
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != "ok" {
			t.Errorf("expected status 'ok', got %q", resp.Status)
		}
		if resp.Version != "1.0.0" {
			t.Errorf("expected version '1.0.0', got %q", resp.Version)
		}
		if resp.Database.Status != "healthy" {
			t.Errorf("expected database 'healthy', got %q", resp.Database.Status)
		}
		if resp.Database.Latency == "" {
			t.Error("expected non-empty latency")
		}
		if resp.Uptime == "" {
			t.Error("expected non-empty uptime")
		}
	})

	t.Run("unhealthy database returns 503", func(t *testing.T) {
		h := NewHandler(&mockPinger{err: errors.New("connection refused")}, "1.0.0")
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", rec.Code)
		}

		var resp Response
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != "degraded" {
			t.Errorf("expected status 'degraded', got %q", resp.Status)
		}
		if resp.Database.Status != "unhealthy" {
			t.Errorf("expected database 'unhealthy', got %q", resp.Database.Status)
		}
	})

	t.Run("response has correct content type", func(t *testing.T) {
		h := NewHandler(&mockPinger{}, "1.0.0")
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
	})
}

func TestHandler_Liveness(t *testing.T) {
	t.Run("returns 200 with ok status", func(t *testing.T) {
		h := NewHandler(&mockPinger{}, "1.0.0")
		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rec := httptest.NewRecorder()

		h.Liveness(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["status"] != "ok" {
			t.Errorf("expected status 'ok', got %q", resp["status"])
		}
	})
}
