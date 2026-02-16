package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"destination-data-aggregation-api/internal/destination"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	svc  *destination.Service
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func (h *Handler) GetByCity(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(chi.URLParam(r, "city"))
	if city == "" {
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	d, err := h.svc.GetByCity(r.Context(), city)
	if err != nil {
		http.Error(w, "destination not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(chi.URLParam(r, "city"))
	if city == "" {
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	d, err := h.svc.Refresh(r.Context(), city)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbOK := h.pool.Ping(ctx) == nil
	redisOK := h.rdb.Ping(ctx).Err() == nil

	status := http.StatusOK
	if !dbOK || !redisOK {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, map[string]bool{
		"postgres": dbOK,
		"redis":    redisOK,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
