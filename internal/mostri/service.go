package mostri

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

func (s *Service) ListMostri(ctx context.Context, filter shared.ListFilter) (*ListMostriResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list mostri", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListMostriResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Mostri:           items,
	}, nil
}

func (s *Service) GetMostro(ctx context.Context, id string) (*Mostro, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get mostri", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrMostroNotFound(id)
	}
	return item, nil
}
