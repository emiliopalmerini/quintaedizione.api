package bastioni

type ListBastioniResponse struct {
	Pagina           int        `json:"pagina"`
	NumeroDiElementi int        `json:"numero-di-elementi"`
	Bastioni         []Bastione `json:"bastioni"`
}
