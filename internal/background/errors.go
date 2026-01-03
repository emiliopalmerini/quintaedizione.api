package background

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrBackgroundNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Background", id)
}
