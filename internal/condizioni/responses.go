package condizioni

type ListCondizioniResponse struct {
	Pagina           int          `json:"pagina"`
	NumeroDiElementi int          `json:"numero-di-elementi"`
	Condizioni       []Condizione `json:"condizioni"`
}
