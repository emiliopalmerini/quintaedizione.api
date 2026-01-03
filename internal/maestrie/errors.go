package maestrie

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrMaestriaNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Maestria", id)
}
