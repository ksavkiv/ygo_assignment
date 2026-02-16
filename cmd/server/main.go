package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"destination-data-aggregation-api/internal/api"
	"destination-data-aggregation-api/internal/cache"
	"destination-data-aggregation-api/internal/config"
	"destination-data-aggregation-api/internal/destination"
	"destination-data-aggregation-api/internal/destination/feed"
	"destination-data-aggregation-api/internal/destination/pipeline"
	"destination-data-aggregation-api/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()

	repo := storage.NewPostgresRepo(pool)
	redisCache := cache.NewRedisCache(rdb)
	fetcher := feed.NewDefaultAPIFetcher()
	svc := destination.NewService(repo, redisCache, fetcher)

	router := api.NewRouter(svc, pool, rdb, cfg.APIToken)

	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Start data feed pipeline in background
	seedCities := []string{"paris", "london", "tokyo"}
	pl := pipeline.NewPipeline(repo, nil, 5*time.Minute, 5*time.Second)
	go pl.Start(ctx, seedCities)

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
