package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type Response struct {
	Status   string         `json:"status"`
	Version  string         `json:"version,omitempty"`
	Uptime   string         `json:"uptime"`
	Database DatabaseStatus `json:"database"`
}

type DatabaseStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
}

type Handler struct {
	db        Pinger
	version   string
	startTime time.Time
}

func NewHandler(db Pinger, version string) *Handler {
	return &Handler{
		db:        db,
		version:   version,
		startTime: time.Now(),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := Response{
		Status:  "ok",
		Version: h.version,
		Uptime:  time.Since(h.startTime).Round(time.Second).String(),
	}

	start := time.Now()
	if err := h.db.PingContext(ctx); err != nil {
		response.Status = "degraded"
		response.Database = DatabaseStatus{
			Status: "unhealthy",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	response.Database = DatabaseStatus{
		Status:  "healthy",
		Latency: time.Since(start).Round(time.Microsecond).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Liveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
