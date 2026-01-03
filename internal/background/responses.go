package background

type ListBackgroundResponse struct {
	Pagina           int          `json:"pagina"`
	NumeroDiElementi int          `json:"numero-di-elementi"`
	Background       []Background `json:"background"`
}
