package api

import (
	"destination-data-aggregation-api/internal/destination"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func NewRouter(svc *destination.Service, pool *pgxpool.Pool, rdb *redis.Client, apiToken string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(BearerAuth(apiToken))

	h := &Handler{svc: svc, pool: pool, rdb: rdb}

	r.Get("/api/v1/health", h.Health)

	r.Route("/api/v1/destinations", func(r chi.Router) {
		r.Get("/{city}", h.GetByCity)
		r.Post("/{city}/refresh", h.Refresh)
	})

	return r
}
