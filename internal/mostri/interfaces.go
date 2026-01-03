package mostri

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]Mostro, int, error)
	GetByID(ctx context.Context, id string) (*Mostro, error)
}
