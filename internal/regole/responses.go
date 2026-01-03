package regole

type ListRegoleResponse struct {
	Pagina           int      `json:"pagina"`
	NumeroDiElementi int      `json:"numero-di-elementi"`
	Regole           []Regola `json:"regole"`
}
