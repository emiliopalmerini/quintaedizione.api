package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
	"github.com/emiliopalmerini/quintaedizione.api/internal/specie"
)

type Handler struct {
	service *specie.Service
}

func NewHandler(service *specie.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListSpecie)
	r.Get("/{id-specie}", h.GetSpecie)

	return r
}

func (h *Handler) ListSpecie(w http.ResponseWriter, r *http.Request) {
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListSpecie(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetSpecie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-specie")
	if err := shared.ValidateID("id-specie", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	item, err := h.service.GetSpecie(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item)
}
