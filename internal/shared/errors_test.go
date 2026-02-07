package shared

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorObject_Constructors(t *testing.T) {
	t.Run("BadRequestError", func(t *testing.T) {
		err := BadRequestError("bad input")
		if len(err.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(err.Errors))
		}
		if err.Errors[0].Code != "BAD_REQUEST" {
			t.Errorf("expected code BAD_REQUEST, got %q", err.Errors[0].Code)
		}
		if err.Errors[0].Detail != "bad input" {
			t.Errorf("expected detail 'bad input', got %q", err.Errors[0].Detail)
		}
	})

	t.Run("NotFoundError", func(t *testing.T) {
		err := NotFoundError("not here")
		if err.Errors[0].Code != "NOT_FOUND" {
			t.Errorf("expected code NOT_FOUND, got %q", err.Errors[0].Code)
		}
	})

	t.Run("InternalServerError", func(t *testing.T) {
		err := InternalServerError("boom")
		if err.Errors[0].Code != "INTERNAL_ERROR" {
			t.Errorf("expected code INTERNAL_ERROR, got %q", err.Errors[0].Code)
		}
	})

	t.Run("UnauthorizedError", func(t *testing.T) {
		err := UnauthorizedError("no access")
		if err.Errors[0].Code != "UNAUTHORIZED" {
			t.Errorf("expected code UNAUTHORIZED, got %q", err.Errors[0].Code)
		}
	})
}

func TestAppError(t *testing.T) {
	t.Run("Error with wrapped error", func(t *testing.T) {
		inner := errors.New("db connection failed")
		appErr := NewAppError(500, InternalServerError("oops"), inner)

		if appErr.Error() != "db connection failed" {
			t.Errorf("expected inner error message, got %q", appErr.Error())
		}
	})

	t.Run("Error without wrapped error", func(t *testing.T) {
		appErr := NewAppError(404, NotFoundError("gone"), nil)

		if appErr.Error() != "Not Found" {
			t.Errorf("expected title as error message, got %q", appErr.Error())
		}
	})

	t.Run("Error with empty response", func(t *testing.T) {
		appErr := &AppError{HTTPStatus: 500, Response: ErrorObject{}}

		if appErr.Error() != "unknown error" {
			t.Errorf("expected 'unknown error', got %q", appErr.Error())
		}
	})

	t.Run("Unwrap returns inner error", func(t *testing.T) {
		inner := errors.New("root cause")
		appErr := NewAppError(500, InternalServerError("oops"), inner)

		if !errors.Is(appErr, inner) {
			t.Error("expected errors.Is to match inner error")
		}
	})

	t.Run("NewBadRequestError", func(t *testing.T) {
		err := NewBadRequestError("invalid", errors.New("parse error"))

		if err.HTTPStatus != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", err.HTTPStatus)
		}
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		err := NewNotFoundError("classe", "barbaro")

		if err.HTTPStatus != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", err.HTTPStatus)
		}
		if err.Response.Errors[0].Detail != "classe with id 'barbaro' not found" {
			t.Errorf("unexpected detail: %q", err.Response.Errors[0].Detail)
		}
	})

	t.Run("NewInternalError", func(t *testing.T) {
		err := NewInternalError(errors.New("boom"))

		if err.HTTPStatus != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", err.HTTPStatus)
		}
		if err.Response.Errors[0].Detail != "an unexpected error occurred" {
			t.Errorf("unexpected detail: %q", err.Response.Errors[0].Detail)
		}
	})
}
