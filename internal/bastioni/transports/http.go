package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/bastioni"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Handler struct {
	service *bastioni.Service
}

func NewHandler(service *bastioni.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListBastioni)
	r.Get("/{id-bastione}", h.GetBastione)

	return r
}

func (h *Handler) ListBastioni(w http.ResponseWriter, r *http.Request) {
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListBastioni(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetBastione(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-bastione")
	if err := shared.ValidateID("id-bastione", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	item, err := h.service.GetBastione(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item)
}
