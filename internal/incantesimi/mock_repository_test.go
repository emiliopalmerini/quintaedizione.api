package incantesimi

import "context"

type MockRepository struct {
	ListFunc    func(ctx context.Context, filter IncantesimiFilter) ([]Incantesimo, int, error)
	GetByIDFunc func(ctx context.Context, id string) (*Incantesimo, error)
}

func (m *MockRepository) List(ctx context.Context, filter IncantesimiFilter) ([]Incantesimo, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Incantesimo, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
