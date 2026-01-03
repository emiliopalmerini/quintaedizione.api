package maestrie

import (
	"context"
	"log/slog"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) ListMaestrie(ctx context.Context, filter shared.ListFilter) (*ListMaestrieResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list maestrie", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListMaestrieResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Maestrie:         items,
	}, nil
}

func (s *Service) GetMaestria(ctx context.Context, id string) (*Maestria, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get maestrie", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrMaestriaNotFound(id)
	}
	return item, nil
}
