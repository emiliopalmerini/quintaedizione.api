package oggetti

type ListOggettiResponse struct {
	Pagina           int       `json:"pagina"`
	NumeroDiElementi int       `json:"numero-di-elementi"`
	Oggetti          []Oggetto `json:"oggetti"`
}
