package classi

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]Classe, int, error)
	GetByID(ctx context.Context, id string) (*Classe, error)
	ListSottoclassi(ctx context.Context, classeID string, filter shared.ListFilter) ([]SottoClasse, int, error)
	GetSottoclasseByID(ctx context.Context, classeID, sottoclasseID string) (*SottoClasse, error)
}
