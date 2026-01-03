package mostri

type ListMostriResponse struct {
	Pagina           int      `json:"pagina"`
	NumeroDiElementi int      `json:"numero-di-elementi"`
	Mostri           []Mostro `json:"mostri"`
}
