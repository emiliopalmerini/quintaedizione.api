package bastioni

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

func (s *Service) ListBastioni(ctx context.Context, filter shared.ListFilter) (*ListBastioniResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list bastioni", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListBastioniResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Bastioni:         items,
	}, nil
}

func (s *Service) GetBastione(ctx context.Context, id string) (*Bastione, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get bastioni", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrBastioneNotFound(id)
	}
	return item, nil
}
