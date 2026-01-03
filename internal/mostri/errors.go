package mostri

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrMostroNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Mostro", id)
}
