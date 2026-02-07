package incantesimi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type IncantesimiFilter struct {
	shared.ListFilter
	Livello             *int32
	ScuolaDiMagia       *string
	Concentrazione      *bool
	Rituale             *bool
	Componenti          []string
	ComponentiMateriali *string
	TempoDiLancio       *string
	Gittata             *string
	Durata              *string
	Classi              []string
}

func NewIncantesimiFilterFromRequest(r *http.Request) (IncantesimiFilter, error) {
	base, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		return IncantesimiFilter{}, err
	}

	filter := IncantesimiFilter{
		ListFilter: base,
	}

	query := r.URL.Query()

	if v := query.Get("livello"); v != "" {
		l, err := strconv.Atoi(v)
		if err != nil {
			return filter, fmt.Errorf("livello must be a valid integer")
		}
		if l < 0 || l > 9 {
			return filter, fmt.Errorf("livello must be between 0 and 9")
		}
		l32 := int32(l)
		filter.Livello = &l32
	}

	if v := query.Get("scuola-di-magia"); v != "" {
		validScuole := map[string]bool{
			"Abiurazione": true, "Divinazione": true, "Evocazione": true,
			"Invocazione": true, "Necromamzia": true, "Illusione": true,
			"Transmutazione": true, "Incantamento": true,
		}
		if !validScuole[v] {
			return filter, fmt.Errorf("scuola-di-magia must be one of: Abiurazione, Divinazione, Evocazione, Invocazione, Necromamzia, Illusione, Transmutazione, Incantamento")
		}
		filter.ScuolaDiMagia = &v
	}

	if v := query.Get("concentrazione"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return filter, fmt.Errorf("concentrazione must be a valid boolean")
		}
		filter.Concentrazione = &b
	}

	if v := query.Get("rituale"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return filter, fmt.Errorf("rituale must be a valid boolean")
		}
		filter.Rituale = &b
	}

	if componenti := query["componenti"]; len(componenti) > 0 {
		validComp := map[string]bool{"V": true, "S": true, "M": true}
		if len(componenti) > 3 {
			return filter, fmt.Errorf("componenti: too many values (max 3)")
		}
		for _, c := range componenti {
			if !validComp[c] {
				return filter, fmt.Errorf("componenti must be one of: V, S, M")
			}
		}
		filter.Componenti = componenti
	}

	if v := query.Get("componenti-materiali"); v != "" {
		if len(v) > 500 {
			return filter, fmt.Errorf("componenti-materiali: value exceeds max length of 500")
		}
		filter.ComponentiMateriali = &v
	}

	if v := query.Get("tempo-di-lancio"); v != "" {
		if len(v) > 255 {
			return filter, fmt.Errorf("tempo-di-lancio: value exceeds max length of 255")
		}
		filter.TempoDiLancio = &v
	}

	if v := query.Get("gittata"); v != "" {
		if len(v) > 255 {
			return filter, fmt.Errorf("gittata: value exceeds max length of 255")
		}
		filter.Gittata = &v
	}

	if v := query.Get("durata"); v != "" {
		if len(v) > 255 {
			return filter, fmt.Errorf("durata: value exceeds max length of 255")
		}
		filter.Durata = &v
	}

	if classi := query["classi"]; len(classi) > 0 {
		if len(classi) > 20 {
			return filter, fmt.Errorf("classi: too many values (max 20)")
		}
		for _, c := range classi {
			if len(c) > 100 {
				return filter, fmt.Errorf("classi: value exceeds max length of 100")
			}
		}
		filter.Classi = classi
	}

	return filter, nil
}
