package transports

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type IncantesimiService interface {
	ListIncantesimi(ctx context.Context, filter incantesimi.IncantesimiFilter) (*incantesimi.ListIncantesimiResponse, error)
	GetIncantesimo(ctx context.Context, id string) (*incantesimi.Incantesimo, error)
}

type Handler struct {
	service IncantesimiService
}

func NewHandler(service IncantesimiService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListIncantesimi)
	r.Get("/{id-incantesimo}", h.GetIncantesimo)

	return r
}

func (h *Handler) ListIncantesimi(w http.ResponseWriter, r *http.Request) {
	filter, err := incantesimi.NewIncantesimiFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListIncantesimi(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetIncantesimo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-incantesimo")
	if err := shared.ValidateID("id-incantesimo", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	result, err := h.service.GetIncantesimo(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, result)
}
