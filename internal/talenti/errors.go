package talenti

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrTalentoNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Talento", id)
}
