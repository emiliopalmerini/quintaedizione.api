package divinita

type ListDivinitaResponse struct {
	Pagina           int        `json:"pagina"`
	NumeroDiElementi int        `json:"numero-di-elementi"`
	Divinita         []Divinita `json:"divinita"`
}
