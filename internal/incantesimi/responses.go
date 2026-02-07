package incantesimi

import "github.com/emiliopalmerini/quintaedizione.api/internal/shared"

type ListIncantesimiResponse struct {
	shared.PaginationMeta
	Incantesimi []Incantesimo `json:"incantesimi"`
}
