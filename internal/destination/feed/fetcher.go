package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"destination-data-aggregation-api/internal/destination"
)

// StubFetcher is a placeholder that returns static data.
// Replace with real external API calls (weather, attractions, etc.).
type StubFetcher struct{}

func NewStubFetcher() *StubFetcher {
	return &StubFetcher{}
}

func (f *StubFetcher) Fetch(_ context.Context, city string) (*destination.Destination, error) {
	return &destination.Destination{
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
	return NewAPIFetcher(GeocodeBaseURL, OpenMeteoBaseURL, RestCountriesBaseURL, AdvisoryBaseURL)
}

func (f *APIFetcher) Fetch(ctx context.Context, city string) (*destination.Destination, error) {
	city = strings.ToLower(city)

	geo, err := FetchGeocode(ctx, f.client, f.geocodeURL, city)
	if err != nil {
		return nil, fmt.Errorf("geocode %s: %w", city, err)
	}

	ch := make(chan FeedResult, 3)
	var wg sync.WaitGroup

	wg.Add(3)
	go func() { defer wg.Done(); ch <- FetchWeather(ctx, f.client, f.weatherURL, city, geo.Latitude, geo.Longitude) }()
	go func() { defer wg.Done(); r := FetchCountry(ctx, f.client, f.countryURL, geo.CountryCode); r.City = city; ch <- r }()
	go func() { defer wg.Done(); ch <- FetchSafety(ctx, f.client, f.safetyURL, city, geo.Country) }()

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

	return &destination.Destination{
		City:      city,
		Country:   geo.Country,
		Latitude:  geo.Latitude,
		Longitude: geo.Longitude,
		Metadata:  metadata,
	}, nil
}
