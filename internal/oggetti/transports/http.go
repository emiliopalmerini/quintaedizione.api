package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/oggetti"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Handler struct {
	service *oggetti.Service
}

func NewHandler(service *oggetti.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListOggetti)
	r.Get("/{id-oggetto}", h.GetOggetto)

	return r
}

func (h *Handler) ListOggetti(w http.ResponseWriter, r *http.Request) {
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListOggetti(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetOggetto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-oggetto")
	if err := shared.ValidateID("id-oggetto", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	item, err := h.service.GetOggetto(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item)
}
