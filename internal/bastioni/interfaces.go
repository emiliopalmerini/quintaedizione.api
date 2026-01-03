package bastioni

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]Bastione, int, error)
	GetByID(ctx context.Context, id string) (*Bastione, error)
}
