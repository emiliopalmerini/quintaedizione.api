package maestrie

type ListMaestrieResponse struct {
	Pagina           int        `json:"pagina"`
	NumeroDiElementi int        `json:"numero-di-elementi"`
	Maestrie         []Maestria `json:"maestrie"`
}
