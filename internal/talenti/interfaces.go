package talenti

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]Talento, int, error)
	GetByID(ctx context.Context, id string) (*Talento, error)
}
