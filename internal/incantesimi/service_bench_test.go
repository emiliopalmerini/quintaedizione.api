package incantesimi

import (
	"context"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

var benchIncantesimi = []Incantesimo{
	{ID: "dardo-incantato_dnd-2024", Nome: "Dardo Incantato", Livello: 1, ScuolaDiMagia: Evocazione},
	{ID: "palla-di-fuoco_dnd-2024", Nome: "Palla di Fuoco", Livello: 3, ScuolaDiMagia: Evocazione},
	{ID: "individuazione-del-magico_dnd-2024", Nome: "Individuazione del Magico", Livello: 1, ScuolaDiMagia: Divinazione},
	{ID: "luce-danzante_dnd-2024", Nome: "Luce Danzante", Livello: 0, ScuolaDiMagia: Illusione},
	{ID: "scudo_dnd-2024", Nome: "Scudo", Livello: 1, ScuolaDiMagia: Abiurazione},
	{ID: "guarigione_dnd-2024", Nome: "Guarigione", Livello: 1, ScuolaDiMagia: Evocazione},
	{ID: "mano-magica_dnd-2024", Nome: "Mano Magica", Livello: 0, ScuolaDiMagia: Invocazione},
	{ID: "fulmine_dnd-2024", Nome: "Fulmine", Livello: 3, ScuolaDiMagia: Evocazione},
	{ID: "immagine-silenziosa_dnd-2024", Nome: "Immagine Silenziosa", Livello: 1, ScuolaDiMagia: Illusione},
	{ID: "tocco-gelido_dnd-2024", Nome: "Tocco Gelido", Livello: 0, ScuolaDiMagia: Necromamzia},
}

func BenchmarkService_ListIncantesimi(b *testing.B) {
	repo := &MockRepository{
		ListFunc: func(_ context.Context, _ IncantesimiFilter) ([]Incantesimo, int, error) {
			return benchIncantesimi, len(benchIncantesimi), nil
		},
	}
	svc := NewService(repo, newTestLogger())
	ctx := context.Background()
	filter := IncantesimiFilter{
		ListFilter: shared.ListFilter{Limit: 20, Offset: 0},
	}

	for b.Loop() {
		_, _ = svc.ListIncantesimi(ctx, filter)
	}
}

func BenchmarkService_GetIncantesimo(b *testing.B) {
	inc := &benchIncantesimi[0]
	repo := &MockRepository{
		GetByIDFunc: func(_ context.Context, _ string) (*Incantesimo, error) {
			return inc, nil
		},
	}
	svc := NewService(repo, newTestLogger())
	ctx := context.Background()

	for b.Loop() {
		_, _ = svc.GetIncantesimo(ctx, "dardo-incantato_dnd-2024")
	}
}
