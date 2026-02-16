package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"destination-data-aggregation-api/internal/destination"
	"destination-data-aggregation-api/internal/destination/feed"
)

// --- mocks ---

type mockRepo struct {
	getByCityFn  func(ctx context.Context, city string) (*destination.Destination, error)
	upsertFn     func(ctx context.Context, d *destination.Destination) error
	listCitiesFn func(ctx context.Context) ([]string, error)
}

func (m *mockRepo) GetByCity(ctx context.Context, city string) (*destination.Destination, error) {
	return m.getByCityFn(ctx, city)
}
func (m *mockRepo) Upsert(ctx context.Context, d *destination.Destination) error {
	return m.upsertFn(ctx, d)
}
func (m *mockRepo) ListCities(ctx context.Context) ([]string, error) {
	if m.listCitiesFn != nil {
		return m.listCitiesFn(ctx)
	}
	return nil, nil
}

// --- tests ---

func TestPipelineListener_BatchesAndFlushes(t *testing.T) {
	var mu sync.Mutex
	upserted := map[string]*destination.Destination{}

	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *destination.Destination) error {
			mu.Lock()
			defer mu.Unlock()
			upserted[d.City] = d
			return nil
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.listen(ctx)
		close(done)
	}()

	// Send 3 FeedResults for "paris": weather, country, safety
	p.ch <- feed.FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)}
	p.ch <- feed.FeedResult{Source: "country", City: "paris", Data: json.RawMessage(`{"name":"France"}`)}
	p.ch <- feed.FeedResult{Source: "safety", City: "paris", Data: json.RawMessage(`{"score":2.8}`)}

	// Store a geocoding result so flush can enrich the destination
	p.geoCache.Store("paris", &feed.GeocodingResult{
		Name:      "Paris",
		Latitude:  48.8566,
		Longitude: 2.3522,
		Country:   "France",
	})

	// Wait for debounce to fire
	time.Sleep(300 * time.Millisecond)

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()

	d, ok := upserted["paris"]
	if !ok {
		t.Fatal("expected upsert to be called for paris")
	}

	var meta map[string]json.RawMessage
	if err := json.Unmarshal(d.Metadata, &meta); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}

	if _, ok := meta["weather"]; !ok {
		t.Error("expected metadata to contain 'weather' key")
	}
	if _, ok := meta["country"]; !ok {
		t.Error("expected metadata to contain 'country' key")
	}
	if _, ok := meta["safety"]; !ok {
		t.Error("expected metadata to contain 'safety' key")
	}
	if len(meta) != 3 {
		t.Errorf("expected 3 metadata keys, got %d", len(meta))
	}
}

func TestPipelineListener_MultipleCities(t *testing.T) {
	var mu sync.Mutex
	upserted := map[string]*destination.Destination{}

	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *destination.Destination) error {
			mu.Lock()
			defer mu.Unlock()
			upserted[d.City] = d
			return nil
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.listen(ctx)
		close(done)
	}()

	// Store geocoding results
	p.geoCache.Store("paris", &feed.GeocodingResult{Name: "Paris", Latitude: 48.8566, Longitude: 2.3522, Country: "France"})
	p.geoCache.Store("london", &feed.GeocodingResult{Name: "London", Latitude: 51.5074, Longitude: -0.1278, Country: "United Kingdom"})

	p.ch <- feed.FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)}
	p.ch <- feed.FeedResult{Source: "weather", City: "london", Data: json.RawMessage(`{"temp":15}`)}

	time.Sleep(300 * time.Millisecond)

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()

	if _, ok := upserted["paris"]; !ok {
		t.Error("expected upsert for paris")
	}
	if _, ok := upserted["london"]; !ok {
		t.Error("expected upsert for london")
	}
	if len(upserted) != 2 {
		t.Errorf("expected 2 upserts, got %d", len(upserted))
	}
}

func TestPipelineListener_SkipsErrors(t *testing.T) {
	var mu sync.Mutex
	upserted := map[string]*destination.Destination{}

	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *destination.Destination) error {
			mu.Lock()
			defer mu.Unlock()
			upserted[d.City] = d
			return nil
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.listen(ctx)
		close(done)
	}()

	p.geoCache.Store("paris", &feed.GeocodingResult{Name: "Paris", Latitude: 48.8566, Longitude: 2.3522, Country: "France"})

	// Send a good weather result and an errored safety result
	p.ch <- feed.FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)}
	p.ch <- feed.FeedResult{Source: "safety", City: "paris", Err: errors.New("api down")}

	time.Sleep(300 * time.Millisecond)

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()

	d, ok := upserted["paris"]
	if !ok {
		t.Fatal("expected upsert for paris")
	}

	var meta map[string]json.RawMessage
	if err := json.Unmarshal(d.Metadata, &meta); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}

	if _, ok := meta["weather"]; !ok {
		t.Error("expected metadata to contain 'weather' key")
	}
	if _, ok := meta["safety"]; ok {
		t.Error("expected metadata NOT to contain 'safety' key (errored result should be skipped)")
	}
	if len(meta) != 1 {
		t.Errorf("expected 1 metadata key, got %d", len(meta))
	}
}

// ---------- pollCity / pollAll ----------

type testServers struct {
	geocode *httptest.Server
	weather *httptest.Server
	country *httptest.Server
	safety  *httptest.Server
}

func newTestServers() *testServers {
	geoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"name":"Paris","latitude":48.8566,"longitude":2.3522,"country":"France","country_code":"FR"}]}`))
	}))

	weatherSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"current_weather":{"temperature":22.5,"windspeed":10.3,"weathercode":1},
			"daily":{"temperature_2m_max":[25.0],"temperature_2m_min":[15.0]}
		}`))
	}))

	countrySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v3.1/all" {
			w.Write([]byte(`[{"capital":["Paris"]}]`))
			return
		}
		w.Write([]byte(`[{
			"name":{"official":"French Republic"},
			"capital":["Paris"],
			"region":"Europe",
			"population":67390000,
			"languages":{"fra":"French"},
			"currencies":{"EUR":{"name":"Euro","symbol":"€"}},
			"flags":{"png":"https://flagcdn.com/w320/fr.png"}
		}]`))
	}))

	safetySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"Title":"France - Level 1: Exercise Normal Precautions","Summary":"Safe","Updated":"2024-06-15"}]`))
	}))

	return &testServers{
		geocode: geoSrv,
		weather: weatherSrv,
		country: countrySrv,
		safety:  safetySrv,
	}
}

func (s *testServers) close() {
	s.geocode.Close()
	s.weather.Close()
	s.country.Close()
	s.safety.Close()
}

func TestPollCity(t *testing.T) {
	servers := newTestServers()
	defer servers.close()

	repo := &mockRepo{
		upsertFn: func(_ context.Context, _ *destination.Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	p.geocodeURL = servers.geocode.URL
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	ctx := context.Background()
	p.pollCity(ctx, "paris")

	// pollCity sends 3 results: weather, country, safety
	results := make([]feed.FeedResult, 0, 3)
	for i := 0; i < 3; i++ {
		select {
		case fr := <-p.ch:
			results = append(results, fr)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for result %d", i+1)
		}
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	sources := map[string]bool{}
	for _, fr := range results {
		if fr.Err != nil {
			t.Errorf("unexpected error for source %q: %v", fr.Source, fr.Err)
		}
		sources[fr.Source] = true
	}

	for _, want := range []string{"weather", "country", "safety"} {
		if !sources[want] {
			t.Errorf("expected result with source %q, not found", want)
		}
	}

	// Verify geocode was cached
	if _, ok := p.geoCache.Load("paris"); !ok {
		t.Error("expected geocode result to be cached")
	}
}

func TestPollCity_GeocodeError(t *testing.T) {
	// Geocode server returns 500, so pollCity should return early with no results
	geoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer geoSrv.Close()

	repo := &mockRepo{
		upsertFn: func(_ context.Context, _ *destination.Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	p.geocodeURL = geoSrv.URL

	ctx := context.Background()
	p.pollCity(ctx, "badcity")

	// Channel should be empty since geocode failed
	select {
	case fr := <-p.ch:
		t.Fatalf("expected no results on channel, got %+v", fr)
	default:
		// expected: nothing on channel
	}
}

func TestPollCity_UsesGeoCache(t *testing.T) {
	// No geocode server needed since we pre-populate the cache
	servers := newTestServers()
	defer servers.close()

	repo := &mockRepo{
		upsertFn: func(_ context.Context, _ *destination.Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	// Pre-populate cache
	p.geoCache.Store("paris", &feed.GeocodingResult{
		Name:        "Paris",
		Latitude:    48.8566,
		Longitude:   2.3522,
		Country:     "France",
		CountryCode: "FR",
	})

	ctx := context.Background()
	p.pollCity(ctx, "paris")

	// Should still get 3 results from the cached geo data
	results := make([]feed.FeedResult, 0, 3)
	for i := 0; i < 3; i++ {
		select {
		case fr := <-p.ch:
			results = append(results, fr)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for result %d", i+1)
		}
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestPollAll(t *testing.T) {
	servers := newTestServers()
	defer servers.close()

	repo := &mockRepo{
		upsertFn: func(_ context.Context, _ *destination.Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	p.pollDelay = 0
	p.geocodeURL = servers.geocode.URL
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	ctx := context.Background()
	p.pollAll(ctx, []string{"paris", "paris"})

	// 2 cities * 3 results each = 6 results
	// (both resolve to same geocode response but that is fine)
	results := make([]feed.FeedResult, 0, 6)
	for i := 0; i < 6; i++ {
		select {
		case fr := <-p.ch:
			results = append(results, fr)
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for result %d (got %d so far)", i+1, len(results))
		}
	}

	if len(results) != 6 {
		t.Fatalf("expected 6 results, got %d", len(results))
	}
}

// ---------- Start ----------

func TestStart_CancelsCleanly(t *testing.T) {
	servers := newTestServers()
	defer servers.close()

	repo := &mockRepo{
		upsertFn: func(_ context.Context, _ *destination.Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 1*time.Hour, 50*time.Millisecond)
	p.pollDelay = 0
	p.geocodeURL = servers.geocode.URL
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.Start(ctx)
		close(done)
	}()

	// Let the initial poll complete, then cancel
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Start returned cleanly
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

func TestStart_PullsCitiesFromAPI(t *testing.T) {
	servers := newTestServers()
	defer servers.close()

	var mu sync.Mutex
	upserted := map[string]bool{}

	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *destination.Destination) error {
			mu.Lock()
			defer mu.Unlock()
			upserted[d.City] = true
			return nil
		},
	}

	p := NewPipeline(repo, nil, 1*time.Hour, 50*time.Millisecond)
	p.pollDelay = 0
	p.geocodeURL = servers.geocode.URL
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.Start(ctx)
		close(done)
	}()

	time.Sleep(300 * time.Millisecond)
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()

	if !upserted["paris"] {
		t.Error("expected pipeline to poll 'paris' from REST Countries API capitals")
	}
}

func TestFlush_UpsertError(t *testing.T) {
	upsertCalls := 0
	repo := &mockRepo{
		upsertFn: func(_ context.Context, _ *destination.Destination) error {
			upsertCalls++
			return errors.New("db error")
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	batch := map[string][]feed.FeedResult{
		"paris": {
			{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)},
		},
	}

	p.flush(context.Background(), batch)

	if upsertCalls != 1 {
		t.Errorf("expected 1 upsert call, got %d", upsertCalls)
	}
}
