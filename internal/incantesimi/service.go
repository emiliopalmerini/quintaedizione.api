package incantesimi

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

func (s *Service) ListIncantesimi(ctx context.Context, filter shared.ListFilter) (*ListIncantesimiResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list incantesimi", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListIncantesimiResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Incantesimi:      items,
	}, nil
}

func (s *Service) GetIncantesimo(ctx context.Context, id string) (*Incantesimo, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get incantesimi", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrIncantesimoNotFound(id)
	}
	return item, nil
}
