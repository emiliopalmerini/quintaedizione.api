package divinita

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]Divinita, int, error)
	GetByID(ctx context.Context, id string) (*Divinita, error)
}
