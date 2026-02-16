# Data Feed Pipeline Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the stub fetcher with a concurrent data feed pipeline that polls 3 external APIs (Open-Meteo weather, REST Countries, travel-advisory.info) every 5 minutes using goroutines, channels, and WaitGroup, with a listener that batches results using a 5-second debounce before persisting.

**Architecture:** A `Pipeline` struct runs alongside the HTTP server. A ticker fires every 5 minutes, spawning fetcher goroutines per city that send `FeedResult` values to a shared channel. A listener goroutine collects results and flushes them as merged metadata after a 5-second debounce window. The pipeline reuses the existing `Repository` interface for persistence.

**Tech Stack:** Go stdlib (`net/http`, `sync`, `time`, `encoding/json`), existing pgx repository

---

### Task 1: FeedResult Type and API Response Types

**Files:**
- Create: `internal/destination/feed.go`

**Step 1: Create the common intermediate type and API response structs**

```go
package destination

import (
	"encoding/json"
	"time"
)

// FeedResult is the common intermediate type produced by all API fetchers.
type FeedResult struct {
	Source    string          // "weather", "country", "safety"
	City     string
	Data     json.RawMessage
	FetchedAt time.Time
	Err      error
}

// GeocodingResult holds coordinates resolved from a city name via Open-Meteo geocoding.
type GeocodingResult struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
	CountryCode string `json:"country_code"`
}

type geocodingResponse struct {
	Results []GeocodingResult `json:"results"`
}

// WeatherData holds the fields we extract from Open-Meteo forecast.
type WeatherData struct {
	CurrentTemp   float64 `json:"current_temp_c"`
	CurrentWind   float64 `json:"current_wind_kmh"`
	WeatherCode   int     `json:"weather_code"`
	DailyMaxTemps []float64 `json:"daily_max_temps"`
	DailyMinTemps []float64 `json:"daily_min_temps"`
}

type openMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
		WeatherCode int     `json:"weathercode"`
	} `json:"current_weather"`
	Daily struct {
		TempMax []float64 `json:"temperature_2m_max"`
		TempMin []float64 `json:"temperature_2m_min"`
	} `json:"daily"`
}

// CountryData holds fields extracted from REST Countries.
type CountryData struct {
	OfficialName string            `json:"official_name"`
	Capital      string            `json:"capital"`
	Region       string            `json:"region"`
	Population   int64             `json:"population"`
	Languages    map[string]string `json:"languages"`
	Currencies   map[string]struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"currencies"`
	FlagURL string `json:"flag_url"`
}

// SafetyData holds fields extracted from travel-advisory.info.
type SafetyData struct {
	Score   float64 `json:"score"`
	Sources int     `json:"sources_active"`
	Message string  `json:"message"`
	Updated string  `json:"updated"`
}

type advisoryAPIResponse struct {
	Data map[string]struct {
		ISOAlpha2 string `json:"iso_alpha2"`
		Name      string `json:"name"`
		Advisory  struct {
			Score   float64 `json:"score"`
			Sources int     `json:"sources_active"`
			Message string  `json:"message"`
			Updated string  `json:"updated"`
		} `json:"advisory"`
	} `json:"data"`
}
```

**Step 2: Run tests to make sure existing tests still pass**

Run: `go test ./internal/destination/... -v`
Expected: All existing tests PASS, new file compiles.

**Step 3: Commit**

```bash
git add internal/destination/feed.go
git commit -m "feat: add FeedResult and API response types for data feed pipeline"
```

---

### Task 2: API Fetcher Functions — Tests

**Files:**
- Create: `internal/destination/feed_test.go`

Write tests for the three fetcher functions using `httptest.Server` to mock external APIs.

**Step 1: Write failing tests for fetchWeather, fetchCountry, fetchSafety**

```go
package destination

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchGeocode_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") != "paris" {
			t.Errorf("expected name=paris, got %s", r.URL.Query().Get("name"))
		}
		json.NewEncoder(w).Encode(geocodingResponse{
			Results: []GeocodingResult{
				{Name: "Paris", Latitude: 48.85, Longitude: 2.35, Country: "France", CountryCode: "FR"},
			},
		})
	}))
	defer srv.Close()

	got, err := fetchGeocode(context.Background(), http.DefaultClient, srv.URL, "paris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CountryCode != "FR" {
		t.Errorf("got country code %q, want FR", got.CountryCode)
	}
	if got.Latitude != 48.85 {
		t.Errorf("got latitude %f, want 48.85", got.Latitude)
	}
}

func TestFetchGeocode_NoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(geocodingResponse{Results: nil})
	}))
	defer srv.Close()

	_, err := fetchGeocode(context.Background(), http.DefaultClient, srv.URL, "atlantis")
	if err == nil {
		t.Fatal("expected error for unknown city, got nil")
	}
}

func TestFetchWeather_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(openMeteoResponse{
			CurrentWeather: struct {
				Temperature float64 `json:"temperature"`
				Windspeed   float64 `json:"windspeed"`
				WeatherCode int     `json:"weathercode"`
			}{Temperature: 15.3, Windspeed: 12.1, WeatherCode: 3},
			Daily: struct {
				TempMax []float64 `json:"temperature_2m_max"`
				TempMin []float64 `json:"temperature_2m_min"`
			}{TempMax: []float64{16.2, 14.8}, TempMin: []float64{8.1, 7.3}},
		})
	}))
	defer srv.Close()

	result := fetchWeather(context.Background(), http.DefaultClient, srv.URL, "paris", 48.85, 2.35)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Source != "weather" {
		t.Errorf("got source %q, want weather", result.Source)
	}
	if result.City != "paris" {
		t.Errorf("got city %q, want paris", result.City)
	}

	var wd WeatherData
	if err := json.Unmarshal(result.Data, &wd); err != nil {
		t.Fatalf("failed to unmarshal weather data: %v", err)
	}
	if wd.CurrentTemp != 15.3 {
		t.Errorf("got temp %f, want 15.3", wd.CurrentTemp)
	}
}

func TestFetchCountry_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3.1/alpha/FR" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`[{
			"name":{"common":"France","official":"French Republic"},
			"capital":["Paris"],
			"region":"Europe",
			"population":67390000,
			"languages":{"fra":"French"},
			"currencies":{"EUR":{"name":"Euro","symbol":"€"}},
			"flags":{"png":"https://flagcdn.com/w320/fr.png"}
		}]`))
	}))
	defer srv.Close()

	result := fetchCountry(context.Background(), http.DefaultClient, srv.URL, "FR")
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Source != "country" {
		t.Errorf("got source %q, want country", result.Source)
	}

	var cd CountryData
	if err := json.Unmarshal(result.Data, &cd); err != nil {
		t.Fatalf("failed to unmarshal country data: %v", err)
	}
	if cd.OfficialName != "French Republic" {
		t.Errorf("got name %q, want French Republic", cd.OfficialName)
	}
	if cd.Capital != "Paris" {
		t.Errorf("got capital %q, want Paris", cd.Capital)
	}
}

func TestFetchSafety_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("countrycode") != "FR" {
			t.Errorf("expected countrycode=FR, got %s", r.URL.Query().Get("countrycode"))
		}
		w.Write([]byte(`{
			"data":{
				"FR":{
					"iso_alpha2":"FR",
					"name":"France",
					"advisory":{"score":2.1,"sources_active":7,"message":"","updated":"2026-02-15"}
				}
			}
		}`))
	}))
	defer srv.Close()

	result := fetchSafety(context.Background(), http.DefaultClient, srv.URL, "paris", "FR")
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Source != "safety" {
		t.Errorf("got source %q, want safety", result.Source)
	}

	var sd SafetyData
	if err := json.Unmarshal(result.Data, &sd); err != nil {
		t.Fatalf("failed to unmarshal safety data: %v", err)
	}
	if sd.Score != 2.1 {
		t.Errorf("got score %f, want 2.1", sd.Score)
	}
}

func TestFetchWeather_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	result := fetchWeather(context.Background(), http.DefaultClient, srv.URL, "paris", 48.85, 2.35)
	if result.Err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestFetchCountry_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	result := fetchCountry(context.Background(), http.DefaultClient, srv.URL, "XX")
	if result.Err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestFetchSafety_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	result := fetchSafety(context.Background(), http.DefaultClient, srv.URL, "paris", "XX")
	if result.Err == nil {
		t.Fatal("expected error for 503 response, got nil")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/destination/... -run TestFetch -v`
Expected: FAIL — functions not defined.

---

### Task 3: API Fetcher Functions — Implementation

**Files:**
- Modify: `internal/destination/feed.go` (append fetcher functions)

**Step 1: Implement fetchGeocode, fetchWeather, fetchCountry, fetchSafety**

Add to `feed.go`:

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func fetchGeocode(ctx context.Context, client *http.Client, baseURL, city string) (*GeocodingResult, error) {
	url := fmt.Sprintf("%s/v1/search?name=%s&count=1", baseURL, city)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode %s: status %d", city, resp.StatusCode)
	}

	var gr geocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return nil, err
	}
	if len(gr.Results) == 0 {
		return nil, fmt.Errorf("geocode %s: no results", city)
	}
	return &gr.Results[0], nil
}

func fetchWeather(ctx context.Context, client *http.Client, baseURL, city string, lat, lon float64) FeedResult {
	url := fmt.Sprintf("%s/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true&daily=temperature_2m_max,temperature_2m_min&timezone=auto&forecast_days=7", baseURL, lat, lon)
	result := FeedResult{Source: "weather", City: city, FetchedAt: time.Now()}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Err = err
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Err = fmt.Errorf("weather API: status %d", resp.StatusCode)
		return result
	}

	var omr openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&omr); err != nil {
		result.Err = err
		return result
	}

	wd := WeatherData{
		CurrentTemp:   omr.CurrentWeather.Temperature,
		CurrentWind:   omr.CurrentWeather.Windspeed,
		WeatherCode:   omr.CurrentWeather.WeatherCode,
		DailyMaxTemps: omr.Daily.TempMax,
		DailyMinTemps: omr.Daily.TempMin,
	}

	data, err := json.Marshal(wd)
	if err != nil {
		result.Err = err
		return result
	}
	result.Data = data
	return result
}

func fetchCountry(ctx context.Context, client *http.Client, baseURL, countryCode string) FeedResult {
	url := fmt.Sprintf("%s/v3.1/alpha/%s", baseURL, strings.ToUpper(countryCode))
	result := FeedResult{Source: "country", City: "", FetchedAt: time.Now()}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Err = err
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Err = fmt.Errorf("country API: status %d", resp.StatusCode)
		return result
	}

	var countries []struct {
		Name struct {
			Common   string `json:"common"`
			Official string `json:"official"`
		} `json:"name"`
		Capital    []string          `json:"capital"`
		Region     string            `json:"region"`
		Population int64             `json:"population"`
		Languages  map[string]string `json:"languages"`
		Currencies map[string]struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"currencies"`
		Flags struct {
			PNG string `json:"png"`
		} `json:"flags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		result.Err = err
		return result
	}
	if len(countries) == 0 {
		result.Err = fmt.Errorf("country API: no results for %s", countryCode)
		return result
	}

	c := countries[0]
	cd := CountryData{
		OfficialName: c.Name.Official,
		Region:       c.Region,
		Population:   c.Population,
		Languages:    c.Languages,
		Currencies:   c.Currencies,
		FlagURL:      c.Flags.PNG,
	}
	if len(c.Capital) > 0 {
		cd.Capital = c.Capital[0]
	}

	data, err := json.Marshal(cd)
	if err != nil {
		result.Err = err
		return result
	}
	result.Data = data
	return result
}

func fetchSafety(ctx context.Context, client *http.Client, baseURL, city, countryCode string) FeedResult {
	url := fmt.Sprintf("%s?countrycode=%s", baseURL, strings.ToUpper(countryCode))
	result := FeedResult{Source: "safety", City: city, FetchedAt: time.Now()}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Err = err
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Err = fmt.Errorf("safety API: status %d", resp.StatusCode)
		return result
	}

	var ar advisoryAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		result.Err = err
		return result
	}

	cc := strings.ToUpper(countryCode)
	entry, ok := ar.Data[cc]
	if !ok {
		result.Err = fmt.Errorf("safety API: no data for %s", cc)
		return result
	}

	sd := SafetyData{
		Score:   entry.Advisory.Score,
		Sources: entry.Advisory.Sources,
		Message: entry.Advisory.Message,
		Updated: entry.Advisory.Updated,
	}

	data, err := json.Marshal(sd)
	if err != nil {
		result.Err = err
		return result
	}
	result.Data = data
	return result
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/destination/... -run TestFetch -v`
Expected: All PASS.

**Step 3: Commit**

```bash
git add internal/destination/feed.go internal/destination/feed_test.go
git commit -m "feat: implement API fetcher functions for weather, country, safety"
```

---

### Task 4: Pipeline Struct and Listener — Tests

**Files:**
- Modify: `internal/destination/feed_test.go` (add pipeline tests)

**Step 1: Write failing tests for pipeline listener debounce and flush**

```go
func TestPipelineListener_BatchesAndFlushes(t *testing.T) {
	var upserted []*Destination
	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *Destination) error {
			upserted = append(upserted, d)
			return nil
		},
		getByCityFn: func(_ context.Context, _ string) (*Destination, error) {
			return nil, errNotFound
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond) // short debounce for test

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.listen(ctx)
		close(done)
	}()

	// Send three results for same city
	p.ch <- FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"current_temp_c":15}`)}
	p.ch <- FeedResult{Source: "country", City: "paris", Data: json.RawMessage(`{"capital":"Paris"}`)}
	p.ch <- FeedResult{Source: "safety", City: "paris", Data: json.RawMessage(`{"score":2.1}`)}

	// Wait for debounce to fire
	time.Sleep(300 * time.Millisecond)
	cancel()
	<-done

	if len(upserted) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(upserted))
	}
	if upserted[0].City != "paris" {
		t.Errorf("got city %q, want paris", upserted[0].City)
	}

	var meta map[string]json.RawMessage
	if err := json.Unmarshal(upserted[0].Metadata, &meta); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	if _, ok := meta["weather"]; !ok {
		t.Error("expected metadata to contain 'weather'")
	}
	if _, ok := meta["country"]; !ok {
		t.Error("expected metadata to contain 'country'")
	}
	if _, ok := meta["safety"]; !ok {
		t.Error("expected metadata to contain 'safety'")
	}
}

func TestPipelineListener_MultipleCities(t *testing.T) {
	var upserted []*Destination
	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *Destination) error {
			upserted = append(upserted, d)
			return nil
		},
		getByCityFn: func(_ context.Context, _ string) (*Destination, error) {
			return nil, errNotFound
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.listen(ctx)
		close(done)
	}()

	p.ch <- FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"current_temp_c":15}`)}
	p.ch <- FeedResult{Source: "weather", City: "london", Data: json.RawMessage(`{"current_temp_c":10}`)}

	time.Sleep(300 * time.Millisecond)
	cancel()
	<-done

	if len(upserted) != 2 {
		t.Fatalf("expected 2 upserts, got %d", len(upserted))
	}
}

func TestPipelineListener_SkipsErrors(t *testing.T) {
	var upserted []*Destination
	repo := &mockRepo{
		upsertFn: func(_ context.Context, d *Destination) error {
			upserted = append(upserted, d)
			return nil
		},
		getByCityFn: func(_ context.Context, _ string) (*Destination, error) {
			return nil, errNotFound
		},
	}

	p := NewPipeline(repo, nil, 5*time.Minute, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.listen(ctx)
		close(done)
	}()

	p.ch <- FeedResult{Source: "weather", City: "paris", Data: json.RawMessage(`{"current_temp_c":15}`)}
	p.ch <- FeedResult{Source: "safety", City: "paris", Err: fmt.Errorf("api down")}

	time.Sleep(300 * time.Millisecond)
	cancel()
	<-done

	if len(upserted) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(upserted))
	}

	var meta map[string]json.RawMessage
	if err := json.Unmarshal(upserted[0].Metadata, &meta); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	if _, ok := meta["weather"]; !ok {
		t.Error("expected metadata to contain 'weather'")
	}
	if _, ok := meta["safety"]; ok {
		t.Error("expected metadata to NOT contain 'safety' (errored)")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/destination/... -run TestPipeline -v`
Expected: FAIL — `NewPipeline` not defined.

---

### Task 5: Pipeline Struct and Listener — Implementation

**Files:**
- Create: `internal/destination/pipeline.go`

**Step 1: Implement Pipeline struct, NewPipeline, listen, flush**

```go
package destination

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Pipeline polls external APIs and feeds results through a channel to a batching listener.
type Pipeline struct {
	repo     Repository
	client   *http.Client
	ch       chan FeedResult
	interval time.Duration
	debounce time.Duration
	geoCache sync.Map // city -> *GeocodingResult
}

func NewPipeline(repo Repository, client *http.Client, interval, debounce time.Duration) *Pipeline {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Pipeline{
		repo:     repo,
		client:   client,
		ch:       make(chan FeedResult, 100),
		interval: interval,
		debounce: debounce,
	}
}

// Start launches the ticker and listener goroutines. Blocks until ctx is cancelled.
func (p *Pipeline) Start(ctx context.Context, cities []string) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		p.listen(ctx)
	}()

	// Initial poll
	p.pollAll(ctx, cities)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.pollAll(ctx, cities)
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

func (p *Pipeline) pollAll(ctx context.Context, cities []string) {
	var wg sync.WaitGroup
	for _, city := range cities {
		wg.Add(1)
		go func(city string) {
			defer wg.Done()
			p.pollCity(ctx, city)
		}(city)
	}
	wg.Wait()
}

const (
	geocodeBaseURL       = "https://geocoding-api.open-meteo.com"
	openMeteoBaseURL     = "https://api.open-meteo.com"
	restCountriesBaseURL = "https://restcountries.com"
	advisoryBaseURL      = "https://www.travel-advisory.info/api"
)

func (p *Pipeline) pollCity(ctx context.Context, city string) {
	// Resolve geocoding (cached in-memory)
	var geo *GeocodingResult
	if cached, ok := p.geoCache.Load(city); ok {
		geo = cached.(*GeocodingResult)
	} else {
		var err error
		geo, err = fetchGeocode(ctx, p.client, geocodeBaseURL, city)
		if err != nil {
			log.Printf("pipeline: geocode %s: %v", city, err)
			return
		}
		p.geoCache.Store(city, geo)
	}

	// Fetch weather first (needs lat/lon)
	weatherResult := fetchWeather(ctx, p.client, openMeteoBaseURL, city, geo.Latitude, geo.Longitude)
	p.ch <- weatherResult

	// Fetch country and safety in parallel
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		r := fetchCountry(ctx, p.client, restCountriesBaseURL, geo.CountryCode)
		r.City = city
		p.ch <- r
	}()
	go func() {
		defer wg.Done()
		p.ch <- fetchSafety(ctx, p.client, advisoryBaseURL, city, geo.CountryCode)
	}()
	wg.Wait()
}

func (p *Pipeline) listen(ctx context.Context) {
	batch := make(map[string][]FeedResult)
	timer := time.NewTimer(p.debounce)
	timer.Stop()

	for {
		select {
		case result := <-p.ch:
			if result.Err != nil {
				log.Printf("pipeline: %s feed error for %s: %v", result.Source, result.City, result.Err)
				continue
			}
			batch[result.City] = append(batch[result.City], result)
			timer.Reset(p.debounce)
		case <-timer.C:
			p.flush(ctx, batch)
			batch = make(map[string][]FeedResult)
		case <-ctx.Done():
			// Drain remaining
			p.flush(ctx, batch)
			return
		}
	}
}

func (p *Pipeline) flush(ctx context.Context, batch map[string][]FeedResult) {
	for city, results := range batch {
		if len(results) == 0 {
			continue
		}
		merged := make(map[string]json.RawMessage)
		for _, r := range results {
			merged[r.Source] = r.Data
		}

		metadata, err := json.Marshal(merged)
		if err != nil {
			log.Printf("pipeline: marshal metadata for %s: %v", city, err)
			continue
		}

		d := &Destination{
			City:     city,
			Metadata: metadata,
		}

		// Try to get existing record to preserve country/lat/lon
		if existing, err := p.repo.GetByCity(ctx, city); err == nil {
			d.Country = existing.Country
			d.Latitude = existing.Latitude
			d.Longitude = existing.Longitude
		}

		// Enrich from geocache
		if geo, ok := p.geoCache.Load(city); ok {
			g := geo.(*GeocodingResult)
			d.Country = g.Country
			d.Latitude = g.Latitude
			d.Longitude = g.Longitude
		}

		if err := p.repo.Upsert(ctx, d); err != nil {
			log.Printf("pipeline: upsert %s: %v", city, err)
		} else {
			log.Printf("pipeline: flushed %s (%d sources)", city, len(results))
		}
	}
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/destination/... -run TestPipeline -v`
Expected: All PASS.

**Step 3: Run all tests**

Run: `go test ./internal/destination/... -v`
Expected: All PASS.

**Step 4: Commit**

```bash
git add internal/destination/pipeline.go internal/destination/feed_test.go
git commit -m "feat: implement pipeline with listener, debounced batching, and flush"
```

---

### Task 6: Wire Pipeline into main.go

**Files:**
- Modify: `cmd/server/main.go`

**Step 1: Add pipeline startup after server launch**

Add after the `go func() { srv.ListenAndServe() }()` block:

```go
	// Start data feed pipeline in background
	seedCities := []string{"paris", "london", "tokyo"}
	pipeline := destination.NewPipeline(repo, nil, 5*time.Minute, 5*time.Second)
	go pipeline.Start(ctx, seedCities)
```

**Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All PASS.

**Step 3: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: wire data feed pipeline into server startup"
```

---

### Task 7: Replace StubFetcher with APIFetcher

**Files:**
- Modify: `internal/destination/fetcher.go`
- Modify: `internal/destination/feed_test.go` (add APIFetcher test)

**Step 1: Write failing test for APIFetcher**

Add to `feed_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/destination/... -run TestAPIFetcher -v`
Expected: FAIL — `NewAPIFetcher` not defined.

**Step 3: Implement APIFetcher in fetcher.go**

Replace `fetcher.go` contents:

```go
package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// StubFetcher is a placeholder that returns static data.
type StubFetcher struct{}

func NewStubFetcher() *StubFetcher {
	return &StubFetcher{}
}

func (f *StubFetcher) Fetch(_ context.Context, city string) (*Destination, error) {
	return &Destination{
		City:     strings.ToLower(city),
		Country:  "unknown",
		Metadata: json.RawMessage(`{"source":"stub"}`),
	}, nil
}

// APIFetcher fetches destination data from Open-Meteo, REST Countries, and travel-advisory.info.
type APIFetcher struct {
	client         *http.Client
	geocodeURL     string
	weatherURL     string
	countryURL     string
	safetyURL      string
}

func NewAPIFetcher(geocodeURL, weatherURL, countryURL, safetyURL string) *APIFetcher {
	return &APIFetcher{
		client:     &http.Client{Timeout: 10 * time.Second},
		geocodeURL: geocodeURL,
		weatherURL: weatherURL,
		countryURL: countryURL,
		safetyURL:  safetyURL,
	}
}

func NewDefaultAPIFetcher() *APIFetcher {
	return NewAPIFetcher(geocodeBaseURL, openMeteoBaseURL, restCountriesBaseURL, advisoryBaseURL)
}

func (f *APIFetcher) Fetch(ctx context.Context, city string) (*Destination, error) {
	city = strings.ToLower(city)

	geo, err := fetchGeocode(ctx, f.client, f.geocodeURL, city)
	if err != nil {
		return nil, fmt.Errorf("geocode %s: %w", city, err)
	}

	ch := make(chan FeedResult, 3)
	var wg sync.WaitGroup

	wg.Add(3)
	go func() { defer wg.Done(); ch <- fetchWeather(ctx, f.client, f.weatherURL, city, geo.Latitude, geo.Longitude) }()
	go func() { defer wg.Done(); r := fetchCountry(ctx, f.client, f.countryURL, geo.CountryCode); r.City = city; ch <- r }()
	go func() { defer wg.Done(); ch <- fetchSafety(ctx, f.client, f.safetyURL, city, geo.CountryCode) }()

	wg.Wait()
	close(ch)

	merged := make(map[string]json.RawMessage)
	for r := range ch {
		if r.Err != nil {
			continue
		}
		merged[r.Source] = r.Data
	}

	metadata, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}

	return &Destination{
		City:      city,
		Country:   geo.Country,
		Latitude:  geo.Latitude,
		Longitude: geo.Longitude,
		Metadata:  metadata,
	}, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/destination/... -v`
Expected: All PASS.

**Step 5: Commit**

```bash
git add internal/destination/fetcher.go internal/destination/feed_test.go
git commit -m "feat: add APIFetcher that aggregates weather, country, safety data"
```

---

### Task 8: Update main.go to use APIFetcher

**Files:**
- Modify: `cmd/server/main.go`

**Step 1: Replace StubFetcher with DefaultAPIFetcher**

Change line 38 from:
```go
	fetcher := destination.NewStubFetcher()
```
to:
```go
	fetcher := destination.NewDefaultAPIFetcher()
```

**Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All PASS.

**Step 3: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: switch from StubFetcher to APIFetcher for live data"
```

---

### Task 9: Coverage Check

**Step 1: Run coverage**

Run: `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`
Expected: All packages >80%.

**Step 2: If below 80%, add missing test cases and re-run.**

---
