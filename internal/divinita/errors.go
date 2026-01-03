package divinita

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrDivinitaNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Divinita", id)
}
