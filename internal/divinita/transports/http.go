package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/divinita"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Handler struct {
	service *divinita.Service
}

func NewHandler(service *divinita.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListDivinita)
	r.Get("/{id-divinita}", h.GetDivinita)

	return r
}

func (h *Handler) ListDivinita(w http.ResponseWriter, r *http.Request) {
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListDivinita(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetDivinita(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-divinita")
	if err := shared.ValidateID("id-divinita", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	item, err := h.service.GetDivinita(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item)
}
