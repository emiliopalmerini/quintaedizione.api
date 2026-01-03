package classi

type ListClassiResponse struct {
	Pagina           int      `json:"pagina"`
	NumeroDiElementi int      `json:"numero-di-elementi"`
	Classi           []Classe `json:"classi"`
}

type ListSottoclassiResponse struct {
	Pagina           int           `json:"pagina"`
	NumeroDiElementi int           `json:"numero-di-elementi"`
	Sottoclassi      []SottoClasse `json:"sottoclassi"`
}
