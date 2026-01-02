package classi

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type MockRepository struct {
	ListFunc             func(ctx context.Context, filter shared.ListFilter) ([]Classe, int, error)
	GetByIDFunc          func(ctx context.Context, id string) (*Classe, error)
	ListSottoclassiFunc  func(ctx context.Context, classeID string, filter shared.ListFilter) ([]SottoClasse, int, error)
	GetSottoclasseByIDFunc func(ctx context.Context, classeID, sottoclasseID string) (*SottoClasse, error)
}

func (m *MockRepository) List(ctx context.Context, filter shared.ListFilter) ([]Classe, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Classe, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) ListSottoclassi(ctx context.Context, classeID string, filter shared.ListFilter) ([]SottoClasse, int, error) {
	if m.ListSottoclassiFunc != nil {
		return m.ListSottoclassiFunc(ctx, classeID, filter)
	}
	return nil, 0, nil
}

func (m *MockRepository) GetSottoclasseByID(ctx context.Context, classeID, sottoclasseID string) (*SottoClasse, error) {
	if m.GetSottoclasseByIDFunc != nil {
		return m.GetSottoclasseByIDFunc(ctx, classeID, sottoclasseID)
	}
	return nil, nil
}
