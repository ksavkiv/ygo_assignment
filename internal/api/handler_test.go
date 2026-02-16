package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"destination-data-aggregation-api/internal/destination"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// --- mock service dependencies ---

type stubRepo struct {
	data map[string]*destination.Destination
}

func (s *stubRepo) GetByCity(_ context.Context, city string) (*destination.Destination, error) {
	if d, ok := s.data[city]; ok {
		return d, nil
	}
	return nil, errors.New("not found")
}
func (s *stubRepo) Upsert(_ context.Context, d *destination.Destination) error {
	s.data[d.City] = d
	return nil
}

type stubCache struct{}

func (s *stubCache) Get(_ context.Context, _ string) (*destination.Destination, error) {
	return nil, errors.New("miss")
}
func (s *stubCache) Set(_ context.Context, _ string, _ *destination.Destination) error { return nil }
func (s *stubCache) Delete(_ context.Context, _ string) error                          { return nil }

type stubFetcher struct{}

func (s *stubFetcher) Fetch(_ context.Context, city string) (*destination.Destination, error) {
	return &destination.Destination{City: city, Country: "test-country"}, nil
}

type failFetcher struct{}

func (f *failFetcher) Fetch(_ context.Context, _ string) (*destination.Destination, error) {
	return nil, errors.New("external api down")
}

func newTestHandler(repo destination.Repository, cache destination.Cache, fetcher destination.Fetcher) *Handler {
	svc := destination.NewService(repo, cache, fetcher)
	return &Handler{svc: svc}
}

func withChiURLParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// --- GetByCity tests ---

func TestGetByCity_Found(t *testing.T) {
	repo := &stubRepo{data: map[string]*destination.Destination{
		"paris": {City: "paris", Country: "france"},
	}}
	h := newTestHandler(repo, &stubCache{}, &stubFetcher{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/destinations/paris", nil)
	req = withChiURLParam(req, "city", "paris")
	w := httptest.NewRecorder()

	h.GetByCity(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("got content-type %q, want application/json", ct)
	}
}

func TestGetByCity_NotFound(t *testing.T) {
	repo := &stubRepo{data: map[string]*destination.Destination{}}
	h := newTestHandler(repo, &stubCache{}, &stubFetcher{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/destinations/atlantis", nil)
	req = withChiURLParam(req, "city", "atlantis")
	w := httptest.NewRecorder()

	h.GetByCity(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- Refresh tests ---

func TestRefresh_Success(t *testing.T) {
	repo := &stubRepo{data: map[string]*destination.Destination{}}
	h := newTestHandler(repo, &stubCache{}, &stubFetcher{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/destinations/tokyo/refresh", nil)
	req = withChiURLParam(req, "city", "tokyo")
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRefresh_FetcherError(t *testing.T) {
	repo := &stubRepo{data: map[string]*destination.Destination{}}
	h := newTestHandler(repo, &stubCache{}, &failFetcher{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/destinations/tokyo/refresh", nil)
	req = withChiURLParam(req, "city", "tokyo")
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// --- Health tests ---

func TestHealth_AllUp(t *testing.T) {
	pool, _ := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable")
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer pool.Close()
	defer rdb.Close()

	h := &Handler{pool: pool, rdb: rdb}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHealth_RedisDown(t *testing.T) {
	pool, _ := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable")
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:9999"})
	defer pool.Close()
	defer rdb.Close()

	h := &Handler{pool: pool, rdb: rdb}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("got status %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}
