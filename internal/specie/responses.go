package specie

type ListSpecieResponse struct {
	Pagina           int      `json:"pagina"`
	NumeroDiElementi int      `json:"numero-di-elementi"`
	Specie           []Specie `json:"specie"`
}
