package linguaggi

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrLinguaggioNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Linguaggio", id)
}
