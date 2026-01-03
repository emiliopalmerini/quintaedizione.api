package background

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

func (s *Service) ListBackground(ctx context.Context, filter shared.ListFilter) (*ListBackgroundResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list background", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListBackgroundResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Background:       items,
	}, nil
}

func (s *Service) GetBackground(ctx context.Context, id string) (*Background, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get background", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrBackgroundNotFound(id)
	}
	return item, nil
}
