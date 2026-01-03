package linguaggi

type ListLinguaggiResponse struct {
	Pagina           int          `json:"pagina"`
	NumeroDiElementi int          `json:"numero-di-elementi"`
	Linguaggi        []Linguaggio `json:"linguaggi"`
}
