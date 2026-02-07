package shared

import (
	"errors"
	"fmt"
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

type listFilterRequest struct {
	Nome   string `validate:"max=100"`
	Sort   string `validate:"omitempty,oneof=asc desc"`
	Limit  int    `validate:"min=1,max=100"`
	Offset int    `validate:"min=0"`
}

func NewListFilterFromRequest(r *http.Request) (ListFilter, error) {
	query := r.URL.Query()

	filter := ListFilter{
		Sort:   SortAsc,
		Limit:  DefaultLimit,
		Offset: 0,
	}

	req := listFilterRequest{
		Nome:   query.Get("nome"),
		Sort:   query.Get("sort"),
		Limit:  DefaultLimit,
		Offset: 0,
	}

	if limit := query.Get("$limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return filter, fmt.Errorf("$limit must be a valid integer")
		}
		req.Limit = l
	}

	if offset := query.Get("$offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return filter, fmt.Errorf("$offset must be a valid integer")
		}
		req.Offset = o
	}

	if err := ValidateStruct(req); err != nil {
		errs := FormatValidationErrors(err)
		if len(errs) > 0 {
			return filter, errors.New(errs[0])
		}
		return filter, err
	}

	if req.Nome != "" {
		filter.Nome = &req.Nome
	}

	if docs := query["documentazione-di-riferimento"]; len(docs) > 0 {
		if len(docs) > 10 {
			return filter, fmt.Errorf("documentazione-di-riferimento: too many values (max 10)")
		}
		for _, d := range docs {
			if len(d) > 100 {
				return filter, fmt.Errorf("documentazione-di-riferimento: value exceeds max length of 100")
			}
		}
		filter.DocumentazioneDiRiferimento = docs
	}

	if req.Sort == "desc" {
		filter.Sort = SortDesc
	}

	filter.Limit = req.Limit
	filter.Offset = req.Offset

	return filter, nil
}

type PaginationMeta struct {
	Pagina           int `json:"pagina"`
	NumeroDiElementi int `json:"numero-di-elementi"`
}

func (f ListFilter) Page() int {
	if f.Limit == 0 {
		return 0
	}
	return (f.Offset / f.Limit) + 1
}
