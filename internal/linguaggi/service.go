package linguaggi

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

func (s *Service) ListLinguaggi(ctx context.Context, filter shared.ListFilter) (*ListLinguaggiResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list linguaggi", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListLinguaggiResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Linguaggi:        items,
	}, nil
}

func (s *Service) GetLinguaggio(ctx context.Context, id string) (*Linguaggio, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get linguaggi", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrLinguaggioNotFound(id)
	}
	return item, nil
}
