package incantesimi

import "context"

type Repository interface {
	List(ctx context.Context, filter IncantesimiFilter) ([]Incantesimo, int, error)
	GetByID(ctx context.Context, id string) (*Incantesimo, error)
}
