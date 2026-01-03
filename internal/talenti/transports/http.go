package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
	"github.com/emiliopalmerini/quintaedizione.api/internal/talenti"
)

type Handler struct {
	service *talenti.Service
}

func NewHandler(service *talenti.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListTalenti)
	r.Get("/{id-talento}", h.GetTalento)

	return r
}

func (h *Handler) ListTalenti(w http.ResponseWriter, r *http.Request) {
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListTalenti(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetTalento(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-talento")
	if err := shared.ValidateID("id-talento", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	item, err := h.service.GetTalento(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item)
}
