package divinita

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

func (s *Service) ListDivinita(ctx context.Context, filter shared.ListFilter) (*ListDivinitaResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list divinita", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListDivinitaResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Divinita:         items,
	}, nil
}

func (s *Service) GetDivinita(ctx context.Context, id string) (*Divinita, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get divinita", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrDivinitaNotFound(id)
	}
	return item, nil
}
