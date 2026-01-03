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
		slog.Error("request error",
			"status", appErr.HTTPStatus,
			"code", appErr.Response.Errors[0].Code,
			"detail", appErr.Response.Errors[0].Detail,
			"error", appErr.Err,
		)
		WriteJSON(w, appErr.HTTPStatus, appErr.Response)
		return
	}

	slog.Error("unexpected error", "error", err)
	WriteJSON(w, http.StatusInternalServerError, InternalServerError("an unexpected error occurred"))
}
