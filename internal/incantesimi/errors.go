package incantesimi

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrIncantesimoNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Incantesimo", id)
}
