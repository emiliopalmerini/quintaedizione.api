package oggetti

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

func (s *Service) ListOggetti(ctx context.Context, filter shared.ListFilter) (*ListOggettiResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list oggetti", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListOggettiResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Oggetti:          items,
	}, nil
}

func (s *Service) GetOggetto(ctx context.Context, id string) (*Oggetto, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get oggetti", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrOggettoNotFound(id)
	}
	return item, nil
}
