package classi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestService_ListClassi(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("success", func(t *testing.T) {
		expectedClassi := []Classe{
			{ID: "barbaro", Nome: "Barbaro", DadoVita: D12},
			{ID: "mago", Nome: "Mago", DadoVita: D6},
		}

		repo := &MockRepository{
			ListFunc: func(_ context.Context, filter shared.ListFilter) ([]Classe, int, error) {
				return expectedClassi, 2, nil
			},
		}

		service := NewService(repo, logger)
		filter := shared.ListFilter{Limit: 20, Offset: 0}

		result, err := service.ListClassi(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.NumeroDiElementi != 2 {
			t.Errorf("expected 2 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Classi) != 2 {
			t.Errorf("expected 2 classi, got %d", len(result.Classi))
		}
		if result.Pagina != 1 {
			t.Errorf("expected page 1, got %d", result.Pagina)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &MockRepository{
			ListFunc: func(_ context.Context, _ shared.ListFilter) ([]Classe, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		service := NewService(repo, logger)
		filter := shared.ListFilter{Limit: 20, Offset: 0}

		_, err := service.ListClassi(ctx, filter)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 500 {
			t.Errorf("expected status 500, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		repo := &MockRepository{
			ListFunc: func(_ context.Context, _ shared.ListFilter) ([]Classe, int, error) {
				return []Classe{}, 0, nil
			},
		}

		service := NewService(repo, logger)
		filter := shared.ListFilter{Limit: 20, Offset: 0}

		result, err := service.ListClassi(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.NumeroDiElementi != 0 {
			t.Errorf("expected 0 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Classi) != 0 {
			t.Errorf("expected 0 classi, got %d", len(result.Classi))
		}
	})
}

func TestService_GetClasse(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("success", func(t *testing.T) {
		expectedClasse := &Classe{
			ID:       "barbaro",
			Nome:     "Barbaro",
			DadoVita: D12,
		}

		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, id string) (*Classe, error) {
				if id == "barbaro" {
					return expectedClasse, nil
				}
				return nil, nil
			},
		}

		service := NewService(repo, logger)

		result, err := service.GetClasse(ctx, "barbaro")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != "barbaro" {
			t.Errorf("expected id 'barbaro', got '%s'", result.ID)
		}
		if result.Nome != "Barbaro" {
			t.Errorf("expected nome 'Barbaro', got '%s'", result.Nome)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return nil, nil
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetClasse(ctx, "nonexistent")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 404 {
			t.Errorf("expected status 404, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetClasse(ctx, "barbaro")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 500 {
			t.Errorf("expected status 500, got %d", appErr.HTTPStatus)
		}
	})
}

func TestService_ListSottoclassi(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("success", func(t *testing.T) {
		parentClasse := &Classe{ID: "barbaro", Nome: "Barbaro"}
		expectedSottoclassi := []SottoClasse{
			{ID: "berserker", Nome: "Berserker", IDClasseAssociata: "barbaro"},
			{ID: "totemico", Nome: "Totemico", IDClasseAssociata: "barbaro"},
		}

		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, id string) (*Classe, error) {
				if id == "barbaro" {
					return parentClasse, nil
				}
				return nil, nil
			},
			ListSottoclassiFunc: func(_ context.Context, classeID string, _ shared.ListFilter) ([]SottoClasse, int, error) {
				if classeID == "barbaro" {
					return expectedSottoclassi, 2, nil
				}
				return nil, 0, nil
			},
		}

		service := NewService(repo, logger)
		filter := shared.ListFilter{Limit: 20, Offset: 0}

		result, err := service.ListSottoclassi(ctx, "barbaro", filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.NumeroDiElementi != 2 {
			t.Errorf("expected 2 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Sottoclassi) != 2 {
			t.Errorf("expected 2 sottoclassi, got %d", len(result.Sottoclassi))
		}
	})

	t.Run("parent classe not found", func(t *testing.T) {
		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return nil, nil
			},
		}

		service := NewService(repo, logger)
		filter := shared.ListFilter{Limit: 20, Offset: 0}

		_, err := service.ListSottoclassi(ctx, "nonexistent", filter)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 404 {
			t.Errorf("expected status 404, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("repository error on get parent", func(t *testing.T) {
		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(repo, logger)
		filter := shared.ListFilter{Limit: 20, Offset: 0}

		_, err := service.ListSottoclassi(ctx, "barbaro", filter)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 500 {
			t.Errorf("expected status 500, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("repository error on list sottoclassi", func(t *testing.T) {
		parentClasse := &Classe{ID: "barbaro", Nome: "Barbaro"}

		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return parentClasse, nil
			},
			ListSottoclassiFunc: func(_ context.Context, _ string, _ shared.ListFilter) ([]SottoClasse, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		service := NewService(repo, logger)
		filter := shared.ListFilter{Limit: 20, Offset: 0}

		_, err := service.ListSottoclassi(ctx, "barbaro", filter)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 500 {
			t.Errorf("expected status 500, got %d", appErr.HTTPStatus)
		}
	})
}

func TestService_GetSottoclasse(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("success", func(t *testing.T) {
		parentClasse := &Classe{ID: "barbaro", Nome: "Barbaro"}
		expectedSottoclasse := &SottoClasse{
			ID:                "berserker",
			Nome:              "Berserker",
			IDClasseAssociata: "barbaro",
		}

		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, id string) (*Classe, error) {
				if id == "barbaro" {
					return parentClasse, nil
				}
				return nil, nil
			},
			GetSottoclasseByIDFunc: func(_ context.Context, classeID, sottoclasseID string) (*SottoClasse, error) {
				if classeID == "barbaro" && sottoclasseID == "berserker" {
					return expectedSottoclasse, nil
				}
				return nil, nil
			},
		}

		service := NewService(repo, logger)

		result, err := service.GetSottoclasse(ctx, "barbaro", "berserker")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != "berserker" {
			t.Errorf("expected id 'berserker', got '%s'", result.ID)
		}
		if result.Nome != "Berserker" {
			t.Errorf("expected nome 'Berserker', got '%s'", result.Nome)
		}
	})

	t.Run("parent classe not found", func(t *testing.T) {
		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return nil, nil
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetSottoclasse(ctx, "nonexistent", "berserker")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 404 {
			t.Errorf("expected status 404, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("sottoclasse not found", func(t *testing.T) {
		parentClasse := &Classe{ID: "barbaro", Nome: "Barbaro"}

		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return parentClasse, nil
			},
			GetSottoclasseByIDFunc: func(_ context.Context, _, _ string) (*SottoClasse, error) {
				return nil, nil
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetSottoclasse(ctx, "barbaro", "nonexistent")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 404 {
			t.Errorf("expected status 404, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("repository error on get parent", func(t *testing.T) {
		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetSottoclasse(ctx, "barbaro", "berserker")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 500 {
			t.Errorf("expected status 500, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("repository error on get sottoclasse", func(t *testing.T) {
		parentClasse := &Classe{ID: "barbaro", Nome: "Barbaro"}

		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
				return parentClasse, nil
			},
			GetSottoclasseByIDFunc: func(_ context.Context, _, _ string) (*SottoClasse, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetSottoclasse(ctx, "barbaro", "berserker")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *shared.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.HTTPStatus != 500 {
			t.Errorf("expected status 500, got %d", appErr.HTTPStatus)
		}
	})
}
