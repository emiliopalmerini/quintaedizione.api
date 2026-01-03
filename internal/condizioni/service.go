package condizioni

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

func (s *Service) ListCondizioni(ctx context.Context, filter shared.ListFilter) (*ListCondizioniResponse, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list condizioni", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListCondizioniResponse{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		Condizioni:       items,
	}, nil
}

func (s *Service) GetCondizione(ctx context.Context, id string) (*Condizione, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get condizioni", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, ErrCondizioneNotFound(id)
	}
	return item, nil
}
