package specie

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

func (s *Service) ListSpecie(ctx context.Context, filter shared.ListFilter) (*ListSpecieResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list specie", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListSpecieResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Specie:           items,
	}, nil
}

func (s *Service) GetSpecie(ctx context.Context, id string) (*Specie, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get specie", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrSpecieNotFound(id)
	}
	return item, nil
}
