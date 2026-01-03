package regole

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

func (s *Service) ListRegole(ctx context.Context, filter shared.ListFilter) (*ListRegoleResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list regole", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListRegoleResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Regole:           items,
	}, nil
}

func (s *Service) GetRegola(ctx context.Context, id string) (*Regola, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get regole", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrRegolaNotFound(id)
	}
	return item, nil
}
