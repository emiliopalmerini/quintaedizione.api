package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

const APIKeyHeader = "X-API-Key"

func APIKey(apiKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get(APIKeyHeader)
			if key == "" {
				shared.WriteJSON(w, http.StatusUnauthorized, shared.UnauthorizedError("missing API key"))
				return
			}

			if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
				shared.WriteJSON(w, http.StatusUnauthorized, shared.UnauthorizedError("invalid API key"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
