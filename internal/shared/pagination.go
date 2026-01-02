package shared

import (
	"net/http"
	"strconv"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type ListFilter struct {
	Nome                        *string
	DocumentazioneDiRiferimento []string
	Sort                        SortOrder
	Limit                       int
	Offset                      int
}

func NewListFilterFromRequest(r *http.Request) ListFilter {
	query := r.URL.Query()

	filter := ListFilter{
		Sort:   SortAsc,
		Limit:  DefaultLimit,
		Offset: 0,
	}

	if nome := query.Get("nome"); nome != "" {
		filter.Nome = &nome
	}

	if docs := query["documentazione-di-riferimento"]; len(docs) > 0 {
		filter.DocumentazioneDiRiferimento = docs
	}

	if sort := query.Get("sort"); sort == "desc" {
		filter.Sort = SortDesc
	}

	if limit := query.Get("$limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			if l > MaxLimit {
				l = MaxLimit
			}
			filter.Limit = l
		}
	}

	if offset := query.Get("$offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	return filter
}

func (f ListFilter) Page() int {
	if f.Limit == 0 {
		return 0
	}
	return (f.Offset / f.Limit) + 1
}
