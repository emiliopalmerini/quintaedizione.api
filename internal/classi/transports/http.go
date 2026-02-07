package transports

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type ClassiService interface {
	ListClassi(ctx context.Context, filter shared.ListFilter) (*classi.ListClassiResponse, error)
	GetClasse(ctx context.Context, id string) (*classi.Classe, error)
	ListSottoclassi(ctx context.Context, classeID string, filter shared.ListFilter) (*classi.ListSottoclassiResponse, error)
	GetSottoclasse(ctx context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error)
}

type Handler struct {
	service ClassiService
}

func NewHandler(service ClassiService) *Handler {
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
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListClassi(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetClasse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id-classe")
	if err := shared.ValidateID("id-classe", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	classe, err := h.service.GetClasse(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, classe)
}

func (h *Handler) ListSottoclassi(w http.ResponseWriter, r *http.Request) {
	classeID := chi.URLParam(r, "id-classe")
	if err := shared.ValidateID("id-classe", classeID); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.ListSottoclassi(r.Context(), classeID, filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetSottoclasse(w http.ResponseWriter, r *http.Request) {
	classeID := chi.URLParam(r, "id-classe")
	if err := shared.ValidateID("id-classe", classeID); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	sottoclasseID := chi.URLParam(r, "id-sotto-classe")
	if err := shared.ValidateID("id-sotto-classe", sottoclasseID); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	sottoclasse, err := h.service.GetSottoclasse(r.Context(), classeID, sottoclasseID)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, sottoclasse)
}
