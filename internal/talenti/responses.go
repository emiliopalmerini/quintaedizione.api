package talenti

type ListTalentiResponse struct {
	Pagina           int       `json:"pagina"`
	NumeroDiElementi int       `json:"numero-di-elementi"`
	Talenti          []Talento `json:"talenti"`
}
