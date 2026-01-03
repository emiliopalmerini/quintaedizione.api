package condizioni

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]Condizione, int, error)
	GetByID(ctx context.Context, id string) (*Condizione, error)
}
