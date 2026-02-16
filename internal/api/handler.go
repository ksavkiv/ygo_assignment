package api

import (
	"encoding/json"
	"log/slog"
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

func (h *Handler) ListCities(w http.ResponseWriter, r *http.Request) {
	slog.Debug("endpoint called", "method", r.Method, "path", r.URL.Path)

	cities, err := h.svc.ListCities(r.Context())
	if err != nil {
		slog.Error("list cities failed", "error", err)
		http.Error(w, "failed to list cities", http.StatusInternalServerError)
		return
	}
	if cities == nil {
		cities = []string{}
	}

	slog.Debug("response", "method", r.Method, "path", r.URL.Path, "count", len(cities), "status", http.StatusOK)
	writeJSON(w, http.StatusOK, cities)
}

func (h *Handler) GetByCity(w http.ResponseWriter, r *http.Request) {
	slog.Debug("endpoint called", "method", r.Method, "path", r.URL.Path)

	city := strings.ToLower(chi.URLParam(r, "city"))
	if city == "" {
		slog.Debug("response", "method", r.Method, "path", r.URL.Path, "status", http.StatusBadRequest)
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	d, err := h.svc.GetByCity(r.Context(), city)
	if err != nil {
		slog.Debug("response", "method", r.Method, "path", r.URL.Path, "city", city, "status", http.StatusNotFound)
		http.Error(w, "destination not found", http.StatusNotFound)
		return
	}
	slog.Debug("response", "method", r.Method, "path", r.URL.Path, "city", city, "status", http.StatusOK)
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	slog.Debug("endpoint called", "method", r.Method, "path", r.URL.Path)

	city := strings.ToLower(chi.URLParam(r, "city"))
	if city == "" {
		slog.Debug("response", "method", r.Method, "path", r.URL.Path, "status", http.StatusBadRequest)
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	d, err := h.svc.Refresh(r.Context(), city)
	if err != nil {
		slog.Debug("response", "method", r.Method, "path", r.URL.Path, "city", city, "status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Debug("response", "method", r.Method, "path", r.URL.Path, "city", city, "status", http.StatusOK)
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	slog.Debug("endpoint called", "method", r.Method, "path", r.URL.Path)
	ctx := r.Context()

	dbErr := h.pool.Ping(ctx)
	redisErr := h.rdb.Ping(ctx).Err()

	if dbErr != nil {
		slog.Error("health: postgres ping failed", "error", dbErr)
	}
	if redisErr != nil {
		slog.Error("health: redis ping failed", "error", redisErr)
	}

	status := http.StatusOK
	if dbErr != nil || redisErr != nil {
		status = http.StatusServiceUnavailable
	}

	slog.Debug("response", "method", r.Method, "path", r.URL.Path, "status", status)
	writeJSON(w, status, map[string]bool{
		"postgres": dbErr == nil,
		"redis":    redisErr == nil,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
