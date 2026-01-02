package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Handler struct {
	service *classi.Service
}

func NewHandler(service *classi.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListClassi)
	r.Get("/{id-classe}", h.GetClasse)
	r.Get("/{id-classe}/sotto-classi", h.ListSottoclassi)
	r.Get("/{id-classe}/sotto-classi/{id-sotto-classe}", h.GetSottoclasse)

	return r
}

func (h *Handler) ListClassi(w http.ResponseWriter, r *http.Request) {
	filter := shared.NewListFilterFromRequest(r)

	response, err := h.service.ListClassi(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetClasse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-classe")

	classe, err := h.service.GetClasse(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, classe)
}

func (h *Handler) ListSottoclassi(w http.ResponseWriter, r *http.Request) {
	classeID := chi.URLParam(r, "id-classe")
	filter := shared.NewListFilterFromRequest(r)

	response, err := h.service.ListSottoclassi(r.Context(), classeID, filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetSottoclasse(w http.ResponseWriter, r *http.Request) {
	classeID := chi.URLParam(r, "id-classe")
	sottoclasseID := chi.URLParam(r, "id-sotto-classe")

	sottoclasse, err := h.service.GetSottoclasse(r.Context(), classeID, sottoclasseID)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, sottoclasse)
}
