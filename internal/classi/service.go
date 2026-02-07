package classi

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

func (s *Service) ListClassi(ctx context.Context, filter shared.ListFilter) (*ListClassiResponse, error) {
	classi, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list classi", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListClassiResponse{
		PaginationMeta: shared.PaginationMeta{Pagina: filter.Page(), NumeroDiElementi: total},
		Classi:         classi,
	}, nil
}

func (s *Service) GetClasse(ctx context.Context, id string) (*Classe, error) {
	classe, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get classe", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if classe == nil {
		return nil, ErrClasseNotFound(id)
	}
	return classe, nil
}

func (s *Service) verifyClasseExists(ctx context.Context, classeID string) error {
	classe, err := s.repo.GetByID(ctx, classeID)
	if err != nil {
		s.logger.Error("failed to verify classe existence", "id", classeID, "error", err)
		return shared.NewInternalError(err)
	}
	if classe == nil {
		return ErrClasseNotFound(classeID)
	}
	return nil
}

func (s *Service) ListSottoclassi(ctx context.Context, classeID string, filter shared.ListFilter) (*ListSottoclassiResponse, error) {
	if err := s.verifyClasseExists(ctx, classeID); err != nil {
		return nil, err
	}

	sottoclassi, total, err := s.repo.ListSottoclassi(ctx, classeID, filter)
	if err != nil {
		s.logger.Error("failed to list sottoclassi", "classeID", classeID, "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &ListSottoclassiResponse{
		PaginationMeta: shared.PaginationMeta{Pagina: filter.Page(), NumeroDiElementi: total},
		Sottoclassi:    sottoclassi,
	}, nil
}

func (s *Service) GetSottoclasse(ctx context.Context, classeID, sottoclasseID string) (*SottoClasse, error) {
	if err := s.verifyClasseExists(ctx, classeID); err != nil {
		return nil, err
	}

	sottoclasse, err := s.repo.GetSottoclasseByID(ctx, classeID, sottoclasseID)
	if err != nil {
		s.logger.Error("failed to get sottoclasse", "classeID", classeID, "sottoclasseID", sottoclasseID, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if sottoclasse == nil {
		return nil, ErrSottoclasseNotFound(sottoclasseID)
	}
	return sottoclasse, nil
}
