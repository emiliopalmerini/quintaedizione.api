package maestrie

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]Maestria, int, error)
	GetByID(ctx context.Context, id string) (*Maestria, error)
}
