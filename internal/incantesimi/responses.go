package incantesimi

type ListIncantesimiResponse struct {
	Pagina           int           `json:"pagina"`
	NumeroDiElementi int           `json:"numero-di-elementi"`
	Incantesimi           []Incantesimo `json:"incantesimi"`
}
