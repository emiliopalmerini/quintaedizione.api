package shared

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("failed to encode JSON response", "error", err)
		}
	}
}

func WriteError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		attrs := []any{"status", appErr.HTTPStatus, "error", appErr.Err}
		if len(appErr.Response.Errors) > 0 {
			attrs = append(attrs, "code", appErr.Response.Errors[0].Code,
				"detail", appErr.Response.Errors[0].Detail)
		}
		if appErr.HTTPStatus >= 500 {
			slog.Error("request error", attrs...)
		} else {
			slog.Warn("request error", attrs...)
		}
		WriteJSON(w, appErr.HTTPStatus, appErr.Response)
		return
	}

	slog.Error("unexpected error", "error", err)
	WriteJSON(w, http.StatusInternalServerError, InternalServerError("an unexpected error occurred"))
}
