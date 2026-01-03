package talenti

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

func (s *Service) ListTalenti(ctx context.Context, filter shared.ListFilter) (*ListTalentiResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list talenti", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListTalentiResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Talenti:          items,
	}, nil
}

func (s *Service) GetTalento(ctx context.Context, id string) (*Talento, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get talenti", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrTalentoNotFound(id)
	}
	return item, nil
}
