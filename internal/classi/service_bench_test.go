package classi

import (
	"context"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

var benchClassi = []Classe{
	{ID: "barbaro", Nome: "Barbaro", DadoVita: D12, Descrizione: "Un feroce guerriero"},
	{ID: "bardo", Nome: "Bardo", DadoVita: D8, Descrizione: "Un artista magico"},
	{ID: "chierico", Nome: "Chierico", DadoVita: D8, Descrizione: "Un sacerdote combattente"},
	{ID: "druido", Nome: "Druido", DadoVita: D8, Descrizione: "Un custode della natura"},
	{ID: "guerriero", Nome: "Guerriero", DadoVita: D10, Descrizione: "Un maestro delle armi"},
	{ID: "ladro", Nome: "Ladro", DadoVita: D8, Descrizione: "Un esperto furtivo"},
	{ID: "mago", Nome: "Mago", DadoVita: D6, Descrizione: "Un incantatore arcano"},
	{ID: "monaco", Nome: "Monaco", DadoVita: D8, Descrizione: "Un artista marziale"},
	{ID: "paladino", Nome: "Paladino", DadoVita: D10, Descrizione: "Un cavaliere sacro"},
	{ID: "ranger", Nome: "Ranger", DadoVita: D10, Descrizione: "Un esploratore"},
}

var benchSottoclassi = []SottoClasse{
	{ID: "berserker", Nome: "Berserker", IDClasseAssociata: "barbaro"},
	{ID: "totemico", Nome: "Totemico", IDClasseAssociata: "barbaro"},
}

func BenchmarkService_ListClassi(b *testing.B) {
	repo := &MockRepository{
		ListFunc: func(_ context.Context, _ shared.ListFilter) ([]Classe, int, error) {
			return benchClassi, len(benchClassi), nil
		},
	}
	svc := NewService(repo, newTestLogger())
	ctx := context.Background()
	filter := shared.ListFilter{Limit: 20, Offset: 0}

	for b.Loop() {
		_, _ = svc.ListClassi(ctx, filter)
	}
}

func BenchmarkService_GetClasse(b *testing.B) {
	classe := &benchClassi[0]
	repo := &MockRepository{
		GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
			return classe, nil
		},
	}
	svc := NewService(repo, newTestLogger())
	ctx := context.Background()

	for b.Loop() {
		_, _ = svc.GetClasse(ctx, "barbaro")
	}
}

func BenchmarkService_ListSottoclassi(b *testing.B) {
	parentClasse := &benchClassi[0]
	repo := &MockRepository{
		GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
			return parentClasse, nil
		},
		ListSottoclassiFunc: func(_ context.Context, _ string, _ shared.ListFilter) ([]SottoClasse, int, error) {
			return benchSottoclassi, len(benchSottoclassi), nil
		},
	}
	svc := NewService(repo, newTestLogger())
	ctx := context.Background()
	filter := shared.ListFilter{Limit: 20, Offset: 0}

	for b.Loop() {
		_, _ = svc.ListSottoclassi(ctx, "barbaro", filter)
	}
}

func BenchmarkService_GetSottoclasse(b *testing.B) {
	parentClasse := &benchClassi[0]
	sottoclasse := &benchSottoclassi[0]
	repo := &MockRepository{
		GetByIDFunc: func(_ context.Context, _ string) (*Classe, error) {
			return parentClasse, nil
		},
		GetSottoclasseByIDFunc: func(_ context.Context, _, _ string) (*SottoClasse, error) {
			return sottoclasse, nil
		},
	}
	svc := NewService(repo, newTestLogger())
	ctx := context.Background()

	for b.Loop() {
		_, _ = svc.GetSottoclasse(ctx, "barbaro", "berserker")
	}
}
