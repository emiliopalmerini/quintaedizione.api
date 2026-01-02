package classi

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func ErrClasseNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("Classe", id)
}

func ErrSottoclasseNotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("SottoClasse", id)
}
