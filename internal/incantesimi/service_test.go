package incantesimi

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

func TestService_ListIncantesimi(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("success", func(t *testing.T) {
		expectedIncantesimi := []Incantesimo{
			{ID: "dardo-incantato_dnd-2024", Nome: "Dardo Incantato", Livello: 1, ScuolaDiMagia: Evocazione},
			{ID: "palla-di-fuoco_dnd-2024", Nome: "Palla di Fuoco", Livello: 3, ScuolaDiMagia: Evocazione},
		}

		repo := &MockRepository{
			ListFunc: func(_ context.Context, _ IncantesimiFilter) ([]Incantesimo, int, error) {
				return expectedIncantesimi, 2, nil
			},
		}

		service := NewService(repo, logger)
		filter := IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0},
		}

		result, err := service.ListIncantesimi(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.NumeroDiElementi != 2 {
			t.Errorf("expected 2 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Incantesimi) != 2 {
			t.Errorf("expected 2 incantesimi, got %d", len(result.Incantesimi))
		}
		if result.Pagina != 1 {
			t.Errorf("expected page 1, got %d", result.Pagina)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &MockRepository{
			ListFunc: func(_ context.Context, _ IncantesimiFilter) ([]Incantesimo, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		service := NewService(repo, logger)
		filter := IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0},
		}

		_, err := service.ListIncantesimi(ctx, filter)

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
			ListFunc: func(_ context.Context, _ IncantesimiFilter) ([]Incantesimo, int, error) {
				return []Incantesimo{}, 0, nil
			},
		}

		service := NewService(repo, logger)
		filter := IncantesimiFilter{
			ListFilter: shared.ListFilter{Limit: 20, Offset: 0},
		}

		result, err := service.ListIncantesimi(ctx, filter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.NumeroDiElementi != 0 {
			t.Errorf("expected 0 elements, got %d", result.NumeroDiElementi)
		}
		if len(result.Incantesimi) != 0 {
			t.Errorf("expected 0 incantesimi, got %d", len(result.Incantesimi))
		}
	})
}

func TestService_GetIncantesimo(t *testing.T) {
	ctx := context.Background()
	logger := newTestLogger()

	t.Run("success", func(t *testing.T) {
		expected := &Incantesimo{
			ID:            "dardo-incantato_dnd-2024",
			Nome:          "Dardo Incantato",
			Livello:       1,
			ScuolaDiMagia: Evocazione,
		}

		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, id string) (*Incantesimo, error) {
				if id == "dardo-incantato_dnd-2024" {
					return expected, nil
				}
				return nil, nil
			},
		}

		service := NewService(repo, logger)

		result, err := service.GetIncantesimo(ctx, "dardo-incantato_dnd-2024")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != "dardo-incantato_dnd-2024" {
			t.Errorf("expected id 'dardo-incantato_dnd-2024', got '%s'", result.ID)
		}
		if result.Nome != "Dardo Incantato" {
			t.Errorf("expected nome 'Dardo Incantato', got '%s'", result.Nome)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &MockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Incantesimo, error) {
				return nil, nil
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetIncantesimo(ctx, "nonexistent")

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
			GetByIDFunc: func(_ context.Context, _ string) (*Incantesimo, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(repo, logger)

		_, err := service.GetIncantesimo(ctx, "dardo-incantato_dnd-2024")

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

func TestService_NilLogger(t *testing.T) {
	repo := &MockRepository{
		ListFunc: func(_ context.Context, _ IncantesimiFilter) ([]Incantesimo, int, error) {
			return []Incantesimo{}, 0, nil
		},
	}

	service := NewService(repo, nil)

	result, err := service.ListIncantesimi(context.Background(), IncantesimiFilter{
		ListFilter: shared.ListFilter{Limit: 20, Offset: 0},
	})

	if err != nil {
		t.Fatalf("unexpected error with nil logger: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
