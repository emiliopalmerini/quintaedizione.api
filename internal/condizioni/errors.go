package condizioni

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrCondizioneNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Condizione", id)
}
