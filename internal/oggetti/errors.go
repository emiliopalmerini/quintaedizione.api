package oggetti

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrOggettoNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Oggetto", id)
}
