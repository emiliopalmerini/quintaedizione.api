package specie

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrSpecieNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Specie", id)
}
