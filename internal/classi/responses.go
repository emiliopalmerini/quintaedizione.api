package classi

import "github.com/emiliopalmerini/quintaedizione.api/internal/shared"

type ListClassiResponse struct {
	shared.PaginationMeta
	Classi []Classe `json:"classi"`
}

type ListSottoclassiResponse struct {
	shared.PaginationMeta
	Sottoclassi []SottoClasse `json:"sottoclassi"`
}
