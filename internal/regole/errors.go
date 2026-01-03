package regole

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrRegolaNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Regola", id)
}
