package bastioni

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrBastioneNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Bastione", id)
}
