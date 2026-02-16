package destination

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// --- mocks ---

type mockRepo struct {
	getByCityFn func(ctx context.Context, city string) (*Destination, error)
	upsertFn    func(ctx context.Context, d *Destination) error
}

func (m *mockRepo) GetByCity(ctx context.Context, city string) (*Destination, error) {
	return m.getByCityFn(ctx, city)
}
func (m *mockRepo) Upsert(ctx context.Context, d *Destination) error {
	return m.upsertFn(ctx, d)
}

type mockCache struct {
	getFn    func(ctx context.Context, key string) (*Destination, error)
	setFn    func(ctx context.Context, key string, d *Destination) error
	deleteFn func(ctx context.Context, key string) error
}

func (m *mockCache) Get(ctx context.Context, key string) (*Destination, error) {
	return m.getFn(ctx, key)
}
func (m *mockCache) Set(ctx context.Context, key string, d *Destination) error {
	return m.setFn(ctx, key, d)
}
func (m *mockCache) Delete(ctx context.Context, key string) error {
	return m.deleteFn(ctx, key)
}

type mockFetcher struct {
	fetchFn func(ctx context.Context, city string) (*Destination, error)
}

func (m *mockFetcher) Fetch(ctx context.Context, city string) (*Destination, error) {
	return m.fetchFn(ctx, city)
}

var errNotFound = errors.New("not found")

func noopSet(_ context.Context, _ string, _ *Destination) error { return nil }
func noopDelete(_ context.Context, _ string) error              { return nil }

// --- tests ---

func TestGetByCity_CacheHit(t *testing.T) {
	want := &Destination{City: "paris", Country: "france"}

	svc := NewService(
		&mockRepo{},
		&mockCache{
			getFn: func(_ context.Context, key string) (*Destination, error) {
				if key != "destination:paris" {
					t.Fatalf("unexpected cache key: %s", key)
				}
				return want, nil
			},
			setFn:    noopSet,
			deleteFn: noopDelete,
		},
		&mockFetcher{},
	)

	got, err := svc.GetByCity(context.Background(), "paris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.City != want.City {
		t.Errorf("got city %q, want %q", got.City, want.City)
	}
}

func TestGetByCity_CacheMiss_DBHit(t *testing.T) {
	want := &Destination{City: "tokyo", Country: "japan"}
	var cached *Destination

	svc := NewService(
		&mockRepo{
			getByCityFn: func(_ context.Context, city string) (*Destination, error) {
				if city != "tokyo" {
					return nil, errNotFound
				}
				return want, nil
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (*Destination, error) {
				return nil, errNotFound
			},
			setFn: func(_ context.Context, _ string, d *Destination) error {
				cached = d
				return nil
			},
			deleteFn: noopDelete,
		},
		&mockFetcher{},
	)

	got, err := svc.GetByCity(context.Background(), "tokyo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.City != want.City {
		t.Errorf("got city %q, want %q", got.City, want.City)
	}
	if cached == nil {
		t.Error("expected result to be cached")
	}
}

func TestGetByCity_CacheMiss_DBMiss(t *testing.T) {
	svc := NewService(
		&mockRepo{
			getByCityFn: func(_ context.Context, _ string) (*Destination, error) {
				return nil, errNotFound
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (*Destination, error) {
				return nil, errNotFound
			},
			setFn:    noopSet,
			deleteFn: noopDelete,
		},
		&mockFetcher{},
	)

	_, err := svc.GetByCity(context.Background(), "atlantis")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRefresh_Success(t *testing.T) {
	fetched := &Destination{City: "berlin", Country: "germany"}
	var upserted, cached bool

	svc := NewService(
		&mockRepo{
			upsertFn: func(_ context.Context, _ *Destination) error {
				upserted = true
				return nil
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (*Destination, error) {
				return nil, errNotFound
			},
			setFn: func(_ context.Context, key string, _ *Destination) error {
				if key != "destination:berlin" {
					t.Fatalf("unexpected cache key: %s", key)
				}
				cached = true
				return nil
			},
			deleteFn: noopDelete,
		},
		&mockFetcher{
			fetchFn: func(_ context.Context, city string) (*Destination, error) {
				return fetched, nil
			},
		},
	)

	got, err := svc.Refresh(context.Background(), "berlin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.City != "berlin" {
		t.Errorf("got city %q, want berlin", got.City)
	}
	if !upserted {
		t.Error("expected repo upsert to be called")
	}
	if !cached {
		t.Error("expected cache set to be called")
	}
}

func TestRefresh_FetcherError(t *testing.T) {
	svc := NewService(
		&mockRepo{},
		&mockCache{getFn: func(_ context.Context, _ string) (*Destination, error) { return nil, errNotFound }, setFn: noopSet, deleteFn: noopDelete},
		&mockFetcher{
			fetchFn: func(_ context.Context, _ string) (*Destination, error) {
				return nil, errors.New("api down")
			},
		},
	)

	_, err := svc.Refresh(context.Background(), "berlin")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRefresh_UpsertError(t *testing.T) {
	svc := NewService(
		&mockRepo{
			upsertFn: func(_ context.Context, _ *Destination) error {
				return errors.New("db error")
			},
		},
		&mockCache{getFn: func(_ context.Context, _ string) (*Destination, error) { return nil, errNotFound }, setFn: noopSet, deleteFn: noopDelete},
		&mockFetcher{
			fetchFn: func(_ context.Context, _ string) (*Destination, error) {
				return &Destination{City: "berlin"}, nil
			},
		},
	)

	_, err := svc.Refresh(context.Background(), "berlin")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRefresh_PreservesMetadata(t *testing.T) {
	meta := json.RawMessage(`{"weather":{"temp_c":22},"poi":["Eiffel Tower"]}`)
	fetched := &Destination{
		City:     "paris",
		Country:  "france",
		Metadata: meta,
	}
	var stored *Destination

	svc := NewService(
		&mockRepo{
			upsertFn: func(_ context.Context, d *Destination) error {
				stored = d
				return nil
			},
		},
		&mockCache{
			getFn:    func(_ context.Context, _ string) (*Destination, error) { return nil, errNotFound },
			setFn:    noopSet,
			deleteFn: noopDelete,
		},
		&mockFetcher{
			fetchFn: func(_ context.Context, _ string) (*Destination, error) {
				return fetched, nil
			},
		},
	)

	got, err := svc.Refresh(context.Background(), "paris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Metadata == nil {
		t.Fatal("expected Metadata to be set, got nil")
	}

	var parsed map[string]any
	if err := json.Unmarshal(got.Metadata, &parsed); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	if _, ok := parsed["weather"]; !ok {
		t.Error("expected metadata to contain 'weather' key")
	}
	if _, ok := parsed["poi"]; !ok {
		t.Error("expected metadata to contain 'poi' key")
	}

	if stored == nil {
		t.Fatal("expected upsert to be called")
	}
	if stored.Metadata == nil {
		t.Error("expected stored destination to have metadata")
	}
}

func TestGetByCity_ReturnsMetadata(t *testing.T) {
	meta := json.RawMessage(`{"safety":{"score":3.5}}`)
	want := &Destination{City: "tokyo", Country: "japan", Metadata: meta}

	svc := NewService(
		&mockRepo{},
		&mockCache{
			getFn: func(_ context.Context, _ string) (*Destination, error) {
				return want, nil
			},
			setFn:    noopSet,
			deleteFn: noopDelete,
		},
		&mockFetcher{},
	)

	got, err := svc.GetByCity(context.Background(), "tokyo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Metadata == nil {
		t.Fatal("expected Metadata to be set, got nil")
	}

	var parsed map[string]any
	if err := json.Unmarshal(got.Metadata, &parsed); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	if _, ok := parsed["safety"]; !ok {
		t.Error("expected metadata to contain 'safety' key")
	}
}
