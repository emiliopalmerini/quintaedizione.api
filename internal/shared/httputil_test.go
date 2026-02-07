package shared

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Run("writes status and JSON body", func(t *testing.T) {
		rec := httptest.NewRecorder()

		WriteJSON(rec, http.StatusOK, map[string]string{"key": "value"})

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["key"] != "value" {
			t.Errorf("expected key=value, got %q", body["key"])
		}
	})

	t.Run("nil data writes no body", func(t *testing.T) {
		rec := httptest.NewRecorder()

		WriteJSON(rec, http.StatusNoContent, nil)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Errorf("expected empty body, got %q", rec.Body.String())
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("AppError with 4xx", func(t *testing.T) {
		rec := httptest.NewRecorder()
		appErr := NewBadRequestError("invalid input", errors.New("parse"))

		WriteError(rec, appErr)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		var body ErrorObject
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Errors[0].Code != "BAD_REQUEST" {
			t.Errorf("expected code BAD_REQUEST, got %q", body.Errors[0].Code)
		}
	})

	t.Run("AppError with 5xx", func(t *testing.T) {
		rec := httptest.NewRecorder()
		appErr := NewInternalError(errors.New("db down"))

		WriteError(rec, appErr)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("non-AppError returns 500", func(t *testing.T) {
		rec := httptest.NewRecorder()

		WriteError(rec, errors.New("unexpected"))

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var body ErrorObject
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Errors[0].Code != "INTERNAL_ERROR" {
			t.Errorf("expected code INTERNAL_ERROR, got %q", body.Errors[0].Code)
		}
	})
}
