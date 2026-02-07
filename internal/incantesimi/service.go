package incantesimi

import (
	"context"
	"io"
	"log/slog"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) ListIncantesimi(ctx context.Context, filter IncantesimiFilter) (*ListIncantesimiResponse, error) {
	incantesimi, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list incantesimi", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListIncantesimiResponse{
		PaginationMeta: shared.PaginationMeta{Pagina: filter.Page(), NumeroDiElementi: total},
		Incantesimi:    incantesimi,
	}, nil
}

func (s *Service) GetIncantesimo(ctx context.Context, id string) (*Incantesimo, error) {
	incantesimo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get incantesimo", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if incantesimo == nil {
		return nil, ErrIncantesimoNotFound(id)
	}
	return incantesimo, nil
}
