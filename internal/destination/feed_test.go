package destination

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// ---------- fetchGeocode ----------

func TestFetchGeocode_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "paris" {
			t.Fatalf("expected name=paris, got %s", got)
		}
		if got := r.URL.Query().Get("count"); got != "1" {
			t.Fatalf("expected count=1, got %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"name":"Paris","latitude":48.8566,"longitude":2.3522,"country":"France","country_code":"FR"}]}`))
	}))
	defer ts.Close()

	result, err := fetchGeocode(context.Background(), ts.Client(), ts.URL, "paris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Paris" {
		t.Errorf("got name %q, want %q", result.Name, "Paris")
	}
	if result.Latitude != 48.8566 {
		t.Errorf("got latitude %f, want 48.8566", result.Latitude)
	}
	if result.Longitude != 2.3522 {
		t.Errorf("got longitude %f, want 2.3522", result.Longitude)
	}
	if result.CountryCode != "FR" {
		t.Errorf("got country_code %q, want %q", result.CountryCode, "FR")
	}
}

func TestFetchGeocode_NoResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	defer ts.Close()

	_, err := fetchGeocode(context.Background(), ts.Client(), ts.URL, "xyznotacity")
	if err == nil {
		t.Fatal("expected error for empty results, got nil")
	}
}

// ---------- fetchWeather ----------

func TestFetchWeather_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/forecast" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("latitude") != "48.856600" {
			t.Fatalf("unexpected latitude: %s", q.Get("latitude"))
		}
		if q.Get("longitude") != "2.352200" {
			t.Fatalf("unexpected longitude: %s", q.Get("longitude"))
		}
		if q.Get("current_weather") != "true" {
			t.Fatalf("expected current_weather=true")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"current_weather":{"temperature":22.5,"windspeed":10.3,"weathercode":1},
			"daily":{"temperature_2m_max":[25.0,26.1,24.3],"temperature_2m_min":[15.0,16.2,14.8]}
		}`))
	}))
	defer ts.Close()

	fr := fetchWeather(context.Background(), ts.Client(), ts.URL, "paris", 48.8566, 2.3522)
	if fr.Err != nil {
		t.Fatalf("unexpected error: %v", fr.Err)
	}
	if fr.Source != "weather" {
		t.Errorf("got source %q, want %q", fr.Source, "weather")
	}
	if fr.City != "paris" {
		t.Errorf("got city %q, want %q", fr.City, "paris")
	}
	if fr.FetchedAt.IsZero() {
		t.Error("expected FetchedAt to be set")
	}

	var wd WeatherData
	if err := json.Unmarshal(fr.Data, &wd); err != nil {
		t.Fatalf("failed to unmarshal WeatherData: %v", err)
	}
	if wd.CurrentTemp != 22.5 {
		t.Errorf("got current_temp %f, want 22.5", wd.CurrentTemp)
	}
	if wd.CurrentWind != 10.3 {
		t.Errorf("got current_wind %f, want 10.3", wd.CurrentWind)
	}
	if wd.WeatherCode != 1 {
		t.Errorf("got weather_code %d, want 1", wd.WeatherCode)
	}
	if len(wd.DailyMaxTemps) != 3 {
		t.Errorf("got %d daily max temps, want 3", len(wd.DailyMaxTemps))
	}
	if len(wd.DailyMinTemps) != 3 {
		t.Errorf("got %d daily min temps, want 3", len(wd.DailyMinTemps))
	}
}

func TestFetchWeather_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	fr := fetchWeather(context.Background(), ts.Client(), ts.URL, "paris", 48.8566, 2.3522)
	if fr.Err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
	if fr.Source != "weather" {
		t.Errorf("got source %q, want %q", fr.Source, "weather")
	}
}

// ---------- fetchCountry ----------

func TestFetchCountry_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3.1/alpha/FR" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{
			"name":{"official":"French Republic"},
			"capital":["Paris"],
			"region":"Europe",
			"population":67390000,
			"languages":{"fra":"French"},
			"currencies":{"EUR":{"name":"Euro","symbol":"\u20ac"}},
			"flags":{"png":"https://flagcdn.com/w320/fr.png"}
		}]`))
	}))
	defer ts.Close()

	fr := fetchCountry(context.Background(), ts.Client(), ts.URL, "FR")
	if fr.Err != nil {
		t.Fatalf("unexpected error: %v", fr.Err)
	}
	if fr.Source != "country" {
		t.Errorf("got source %q, want %q", fr.Source, "country")
	}
	if fr.FetchedAt.IsZero() {
		t.Error("expected FetchedAt to be set")
	}

	var cd CountryData
	if err := json.Unmarshal(fr.Data, &cd); err != nil {
		t.Fatalf("failed to unmarshal CountryData: %v", err)
	}
	if cd.OfficialName != "French Republic" {
		t.Errorf("got official_name %q, want %q", cd.OfficialName, "French Republic")
	}
	if cd.Capital != "Paris" {
		t.Errorf("got capital %q, want %q", cd.Capital, "Paris")
	}
	if cd.Region != "Europe" {
		t.Errorf("got region %q, want %q", cd.Region, "Europe")
	}
	if cd.Population != 67390000 {
		t.Errorf("got population %d, want 67390000", cd.Population)
	}
	if cd.Languages["fra"] != "French" {
		t.Errorf("got languages %v, want fra=French", cd.Languages)
	}
	if cd.Currencies["EUR"].Name != "Euro" {
		t.Errorf("got currency name %q, want %q", cd.Currencies["EUR"].Name, "Euro")
	}
	if cd.FlagURL != "https://flagcdn.com/w320/fr.png" {
		t.Errorf("got flag_url %q, want %q", cd.FlagURL, "https://flagcdn.com/w320/fr.png")
	}
}

func TestFetchCountry_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	fr := fetchCountry(context.Background(), ts.Client(), ts.URL, "ZZ")
	if fr.Err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if fr.Source != "country" {
		t.Errorf("got source %q, want %q", fr.Source, "country")
	}
}

// ---------- fetchSafety ----------

func TestFetchSafety_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("countrycode"); got != "FR" {
			t.Fatalf("expected countrycode=FR, got %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"api_status":{"reply":{"code":200}},
			"data":{
				"FR":{
					"iso_alpha2":"FR",
					"name":"France",
					"advisory":{
						"score":2.8,
						"sources_active":7,
						"message":"Exercise normal safety precautions",
						"updated":"2024-06-15 10:30:00"
					}
				}
			}
		}`))
	}))
	defer ts.Close()

	fr := fetchSafety(context.Background(), ts.Client(), ts.URL, "paris", "FR")
	if fr.Err != nil {
		t.Fatalf("unexpected error: %v", fr.Err)
	}
	if fr.Source != "safety" {
		t.Errorf("got source %q, want %q", fr.Source, "safety")
	}
	if fr.City != "paris" {
		t.Errorf("got city %q, want %q", fr.City, "paris")
	}
	if fr.FetchedAt.IsZero() {
		t.Error("expected FetchedAt to be set")
	}

	var sd SafetyData
	if err := json.Unmarshal(fr.Data, &sd); err != nil {
		t.Fatalf("failed to unmarshal SafetyData: %v", err)
	}
	if sd.Score != 2.8 {
		t.Errorf("got score %f, want 2.8", sd.Score)
	}
	if sd.Sources != 7 {
		t.Errorf("got sources %d, want 7", sd.Sources)
	}
	if sd.Message != "Exercise normal safety precautions" {
		t.Errorf("got message %q, want %q", sd.Message, "Exercise normal safety precautions")
	}
	if sd.Updated != "2024-06-15 10:30:00" {
		t.Errorf("got updated %q, want %q", sd.Updated, "2024-06-15 10:30:00")
	}
}

func TestFetchSafety_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	fr := fetchSafety(context.Background(), ts.Client(), ts.URL, "paris", "FR")
	if fr.Err == nil {
		t.Fatal("expected error for 503 response, got nil")
	}
	if fr.Source != "safety" {
		t.Errorf("got source %q, want %q", fr.Source, "safety")
	}
}

// ---------- Pipeline tests ----------

func TestPipelineListener_BatchesAndFlushes(t *testing.T) {
	var mu sync.Mutex
	upserted := map[string]*Destination{}

	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *Destination) error {
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
	p.ch <- FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)}
	p.ch <- FeedResult{Source: "country", City: "paris", Data: json.RawMessage(`{"name":"France"}`)}
	p.ch <- FeedResult{Source: "safety", City: "paris", Data: json.RawMessage(`{"score":2.8}`)}

	// Store a geocoding result so flush can enrich the destination
	p.geoCache.Store("paris", &GeocodingResult{
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
	upserted := map[string]*Destination{}

	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *Destination) error {
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
	p.geoCache.Store("paris", &GeocodingResult{Name: "Paris", Latitude: 48.8566, Longitude: 2.3522, Country: "France"})
	p.geoCache.Store("london", &GeocodingResult{Name: "London", Latitude: 51.5074, Longitude: -0.1278, Country: "United Kingdom"})

	p.ch <- FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)}
	p.ch <- FeedResult{Source: "weather", City: "london", Data: json.RawMessage(`{"temp":15}`)}

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

// ---------- APIFetcher ----------

func TestAPIFetcher_Fetch(t *testing.T) {
	geoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(geocodingResponse{
			Results: []GeocodingResult{
				{Name: "Paris", Latitude: 48.85, Longitude: 2.35, Country: "France", CountryCode: "FR"},
			},
		})
	}))
	defer geoSrv.Close()

	weatherSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(openMeteoResponse{
			CurrentWeather: struct {
				Temperature float64 `json:"temperature"`
				Windspeed   float64 `json:"windspeed"`
				WeatherCode int     `json:"weathercode"`
			}{Temperature: 15.3, Windspeed: 12.1, WeatherCode: 3},
			Daily: struct {
				TempMax []float64 `json:"temperature_2m_max"`
				TempMin []float64 `json:"temperature_2m_min"`
			}{TempMax: []float64{16.2}, TempMin: []float64{8.1}},
		})
	}))
	defer weatherSrv.Close()

	countrySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":{"common":"France","official":"French Republic"},"capital":["Paris"],"region":"Europe","population":67390000,"languages":{"fra":"French"},"currencies":{"EUR":{"name":"Euro","symbol":"€"}},"flags":{"png":"https://flagcdn.com/w320/fr.png"}}]`))
	}))
	defer countrySrv.Close()

	safetySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"FR":{"iso_alpha2":"FR","name":"France","advisory":{"score":2.1,"sources_active":7,"message":"","updated":"2026-02-15"}}}}`))
	}))
	defer safetySrv.Close()

	f := NewAPIFetcher(geoSrv.URL, weatherSrv.URL, countrySrv.URL, safetySrv.URL)
	d, err := f.Fetch(context.Background(), "paris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.City != "paris" {
		t.Errorf("got city %q, want paris", d.City)
	}
	if d.Country != "France" {
		t.Errorf("got country %q, want France", d.Country)
	}
	if d.Latitude != 48.85 {
		t.Errorf("got lat %f, want 48.85", d.Latitude)
	}

	var meta map[string]json.RawMessage
	if err := json.Unmarshal(d.Metadata, &meta); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	for _, key := range []string{"weather", "country", "safety"} {
		if _, ok := meta[key]; !ok {
			t.Errorf("expected metadata to contain %q", key)
		}
	}
}

func TestPipelineListener_SkipsErrors(t *testing.T) {
	var mu sync.Mutex
	upserted := map[string]*Destination{}

	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *Destination) error {
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

	p.geoCache.Store("paris", &GeocodingResult{Name: "Paris", Latitude: 48.8566, Longitude: 2.3522, Country: "France"})

	// Send a good weather result and an errored safety result
	p.ch <- FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)}
	p.ch <- FeedResult{Source: "safety", City: "paris", Err: errors.New("api down")}

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

// ---------- StubFetcher ----------

func TestStubFetcher(t *testing.T) {
	f := NewStubFetcher()
	d, err := f.Fetch(context.Background(), "Paris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.City != "paris" {
		t.Errorf("got city %q, want %q (lowercase)", d.City, "paris")
	}
	if d.Country != "unknown" {
		t.Errorf("got country %q, want %q", d.Country, "unknown")
	}
	if d.Metadata == nil {
		t.Fatal("expected metadata to be set, got nil")
	}

	var meta map[string]string
	if err := json.Unmarshal(d.Metadata, &meta); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}
	if meta["source"] != "stub" {
		t.Errorf("got source %q, want %q", meta["source"], "stub")
	}
}

// ---------- NewDefaultAPIFetcher ----------

func TestNewDefaultAPIFetcher(t *testing.T) {
	f := NewDefaultAPIFetcher()
	if f == nil {
		t.Fatal("expected non-nil APIFetcher")
	}
	if f.geocodeURL != geocodeBaseURL {
		t.Errorf("got geocodeURL %q, want %q", f.geocodeURL, geocodeBaseURL)
	}
	if f.weatherURL != openMeteoBaseURL {
		t.Errorf("got weatherURL %q, want %q", f.weatherURL, openMeteoBaseURL)
	}
	if f.countryURL != restCountriesBaseURL {
		t.Errorf("got countryURL %q, want %q", f.countryURL, restCountriesBaseURL)
	}
	if f.safetyURL != advisoryBaseURL {
		t.Errorf("got safetyURL %q, want %q", f.safetyURL, advisoryBaseURL)
	}
}

// ---------- pollCity / pollAll ----------

// testEndpoints creates httptest servers for all 4 external APIs and returns
// their URLs plus a cleanup function.
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
		w.Write([]byte(`{
			"data":{
				"FR":{
					"iso_alpha2":"FR",
					"name":"France",
					"advisory":{"score":2.8,"sources_active":7,"message":"Safe","updated":"2024-06-15"}
				}
			}
		}`))
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
		upsertFn: func(_ context.Context, _ *Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	p.geocodeURL = servers.geocode.URL
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	ctx := context.Background()
	p.pollCity(ctx, "paris")

	// pollCity sends 3 results: weather, country, safety
	results := make([]FeedResult, 0, 3)
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
		upsertFn: func(_ context.Context, _ *Destination) error { return nil },
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
		upsertFn: func(_ context.Context, _ *Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	// Pre-populate cache
	p.geoCache.Store("paris", &GeocodingResult{
		Name:        "Paris",
		Latitude:    48.8566,
		Longitude:   2.3522,
		Country:     "France",
		CountryCode: "FR",
	})

	ctx := context.Background()
	p.pollCity(ctx, "paris")

	// Should still get 3 results from the cached geo data
	results := make([]FeedResult, 0, 3)
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
		upsertFn: func(_ context.Context, _ *Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	p.geocodeURL = servers.geocode.URL
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	ctx := context.Background()
	p.pollAll(ctx, []string{"paris", "paris"})

	// 2 cities * 3 results each = 6 results
	// (both resolve to same geocode response but that is fine)
	results := make([]FeedResult, 0, 6)
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
		upsertFn: func(_ context.Context, _ *Destination) error { return nil },
	}

	p := NewPipeline(repo, nil, 1*time.Hour, 50*time.Millisecond)
	p.geocodeURL = servers.geocode.URL
	p.weatherURL = servers.weather.URL
	p.countryURL = servers.country.URL
	p.safetyURL = servers.safety.URL

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.Start(ctx, []string{"paris"})
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

// ---------- Additional error branch tests ----------

func TestFetchGeocode_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := fetchGeocode(context.Background(), ts.Client(), ts.URL, "paris")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestFetchGeocode_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer ts.Close()

	_, err := fetchGeocode(context.Background(), ts.Client(), ts.URL, "paris")
	if err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
}

func TestFetchWeather_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer ts.Close()

	fr := fetchWeather(context.Background(), ts.Client(), ts.URL, "paris", 48.8566, 2.3522)
	if fr.Err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
	if fr.Source != "weather" {
		t.Errorf("got source %q, want %q", fr.Source, "weather")
	}
}

func TestFetchCountry_EmptyArray(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	fr := fetchCountry(context.Background(), ts.Client(), ts.URL, "ZZ")
	if fr.Err == nil {
		t.Fatal("expected error for empty array, got nil")
	}
	if fr.Source != "country" {
		t.Errorf("got source %q, want %q", fr.Source, "country")
	}
}

func TestFetchCountry_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer ts.Close()

	fr := fetchCountry(context.Background(), ts.Client(), ts.URL, "FR")
	if fr.Err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
	if fr.Source != "country" {
		t.Errorf("got source %q, want %q", fr.Source, "country")
	}
}

func TestFetchSafety_MissingCountryCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Valid JSON but without the requested country code "FR"
		w.Write([]byte(`{"data":{"DE":{"iso_alpha2":"DE","name":"Germany","advisory":{"score":1.5,"sources_active":5,"message":"Safe","updated":"2024-01-01"}}}}`))
	}))
	defer ts.Close()

	fr := fetchSafety(context.Background(), ts.Client(), ts.URL, "paris", "FR")
	if fr.Err == nil {
		t.Fatal("expected error for missing country code, got nil")
	}
	if fr.Source != "safety" {
		t.Errorf("got source %q, want %q", fr.Source, "safety")
	}
}

func TestFetchSafety_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer ts.Close()

	fr := fetchSafety(context.Background(), ts.Client(), ts.URL, "paris", "FR")
	if fr.Err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
	if fr.Source != "safety" {
		t.Errorf("got source %q, want %q", fr.Source, "safety")
	}
}

func TestFetchGeocode_ConnectionError(t *testing.T) {
	// Use a URL that will fail to connect
	_, err := fetchGeocode(context.Background(), &http.Client{Timeout: 100 * time.Millisecond}, "http://127.0.0.1:1", "paris")
	if err == nil {
		t.Fatal("expected error for connection failure, got nil")
	}
}

func TestFetchWeather_ConnectionError(t *testing.T) {
	fr := fetchWeather(context.Background(), &http.Client{Timeout: 100 * time.Millisecond}, "http://127.0.0.1:1", "paris", 48.0, 2.0)
	if fr.Err == nil {
		t.Fatal("expected error for connection failure, got nil")
	}
}

func TestFetchCountry_ConnectionError(t *testing.T) {
	fr := fetchCountry(context.Background(), &http.Client{Timeout: 100 * time.Millisecond}, "http://127.0.0.1:1", "FR")
	if fr.Err == nil {
		t.Fatal("expected error for connection failure, got nil")
	}
}

func TestFetchSafety_ConnectionError(t *testing.T) {
	fr := fetchSafety(context.Background(), &http.Client{Timeout: 100 * time.Millisecond}, "http://127.0.0.1:1", "paris", "FR")
	if fr.Err == nil {
		t.Fatal("expected error for connection failure, got nil")
	}
}

func TestFlush_UpsertError(t *testing.T) {
	upsertCalls := 0
	repo := &mockRepo{
		upsertFn: func(_ context.Context, _ *Destination) error {
			upsertCalls++
			return errors.New("db error")
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)
	batch := map[string][]FeedResult{
		"paris": {
			{Source: "weather", City: "paris", Data: json.RawMessage(`{"temp":22}`)},
		},
	}

	p.flush(context.Background(), batch)

	if upsertCalls != 1 {
		t.Errorf("expected 1 upsert call, got %d", upsertCalls)
	}
}
