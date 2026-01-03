package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/regole"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Handler struct {
	service *regole.Service
}

func NewHandler(service *regole.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListRegole)
	r.Get("/{id-regola}", h.GetRegola)

	return r
}

func (h *Handler) ListRegole(w http.ResponseWriter, r *http.Request) {
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListRegole(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetRegola(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-regola")
	if err := shared.ValidateID("id-regola", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	item, err := h.service.GetRegola(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item)
}
