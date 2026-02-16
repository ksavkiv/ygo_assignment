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
// Replace with real external API calls (weather, attractions, etc.).
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
	client     *http.Client
	geocodeURL string
	weatherURL string
	countryURL string
	safetyURL  string
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
