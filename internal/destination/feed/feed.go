package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	GeocodeBaseURL       = "https://geocoding-api.open-meteo.com"
	OpenMeteoBaseURL     = "https://api.open-meteo.com"
	RestCountriesBaseURL = "https://restcountries.com"
	AdvisoryBaseURL      = "https://cadataapi.state.gov/api/TravelAdvisories"
)

// FeedResult is the common intermediate type produced by all API fetchers.
type FeedResult struct {
	Source    string // "weather", "country", "safety"
	City     string
	Data     json.RawMessage
	FetchedAt time.Time
	Err      error
}

// GeocodingResult holds coordinates resolved from a city name via Open-Meteo geocoding.
type GeocodingResult struct {
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
}

type geocodingResponse struct {
	Results []GeocodingResult `json:"results"`
}

// WeatherData holds the fields we extract from Open-Meteo forecast.
type WeatherData struct {
	CurrentTemp   float64   `json:"current_temp_c"`
	CurrentWind   float64   `json:"current_wind_kmh"`
	WeatherCode   int       `json:"weather_code"`
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

// SafetyData holds fields extracted from the US State Department Travel Advisories API.
type SafetyData struct {
	Score   float64 `json:"score"`
	Sources int     `json:"sources_active"`
	Message string  `json:"message"`
	Updated string  `json:"updated"`
}

type stateDeptAdvisory struct {
	Title   string `json:"Title"`
	Summary string `json:"Summary"`
	Updated string `json:"Updated"`
}

// restCountryItem mirrors the relevant fields of a single REST Countries array element.
type restCountryItem struct {
	Name struct {
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

// ---------- fetcher functions ----------

// FetchGeocode resolves a city name to coordinates using the Open-Meteo geocoding API.
func FetchGeocode(ctx context.Context, client *http.Client, baseURL, city string) (*GeocodingResult, error) {
	url := fmt.Sprintf("%s/v1/search?name=%s&count=1", baseURL, url.QueryEscape(city))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("geocode request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode: unexpected status %d", resp.StatusCode)
	}

	var gr geocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return nil, fmt.Errorf("geocode decode: %w", err)
	}

	if len(gr.Results) == 0 {
		return nil, fmt.Errorf("geocode: no results for %q", city)
	}

	return &gr.Results[0], nil
}

// FetchWeather retrieves the current weather and 7-day forecast from Open-Meteo.
func FetchWeather(ctx context.Context, client *http.Client, baseURL, city string, lat, lon float64) FeedResult {
	fr := FeedResult{Source: "weather", City: city}

	url := fmt.Sprintf(
		"%s/v1/forecast?latitude=%f&longitude=%f&current_weather=true&daily=temperature_2m_max,temperature_2m_min&timezone=auto&forecast_days=7",
		baseURL, lat, lon,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fr.Err = fmt.Errorf("weather request: %w", err)
		return fr
	}

	resp, err := client.Do(req)
	if err != nil {
		fr.Err = fmt.Errorf("weather call: %w", err)
		return fr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fr.Err = fmt.Errorf("weather: unexpected status %d", resp.StatusCode)
		return fr
	}

	var omr openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&omr); err != nil {
		fr.Err = fmt.Errorf("weather decode: %w", err)
		return fr
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
		fr.Err = fmt.Errorf("weather marshal: %w", err)
		return fr
	}

	fr.Data = data
	fr.FetchedAt = time.Now()
	return fr
}

// FetchCountry retrieves country information from REST Countries by alpha code.
func FetchCountry(ctx context.Context, client *http.Client, baseURL, countryCode string) FeedResult {
	fr := FeedResult{Source: "country"}

	url := fmt.Sprintf("%s/v3.1/alpha/%s", baseURL, strings.ToUpper(countryCode))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fr.Err = fmt.Errorf("country request: %w", err)
		return fr
	}

	resp, err := client.Do(req)
	if err != nil {
		fr.Err = fmt.Errorf("country call: %w", err)
		return fr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fr.Err = fmt.Errorf("country: unexpected status %d", resp.StatusCode)
		return fr
	}

	var items []restCountryItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		fr.Err = fmt.Errorf("country decode: %w", err)
		return fr
	}

	if len(items) == 0 {
		fr.Err = fmt.Errorf("country: no results for %q", countryCode)
		return fr
	}

	item := items[0]
	cd := CountryData{
		OfficialName: item.Name.Official,
		Region:       item.Region,
		Population:   item.Population,
		Languages:    item.Languages,
		Currencies:   item.Currencies,
		FlagURL:      item.Flags.PNG,
	}
	if len(item.Capital) > 0 {
		cd.Capital = item.Capital[0]
	}

	data, err := json.Marshal(cd)
	if err != nil {
		fr.Err = fmt.Errorf("country marshal: %w", err)
		return fr
	}

	fr.Data = data
	fr.FetchedAt = time.Now()
	return fr
}

// FetchSafety retrieves travel advisory data from the US State Department API.
func FetchSafety(ctx context.Context, client *http.Client, baseURL, city, countryName string) FeedResult {
	fr := FeedResult{Source: "safety", City: city}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		fr.Err = fmt.Errorf("safety request: %w", err)
		return fr
	}

	resp, err := client.Do(req)
	if err != nil {
		fr.Err = fmt.Errorf("safety call: %w", err)
		return fr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fr.Err = fmt.Errorf("safety: unexpected status %d", resp.StatusCode)
		return fr
	}

	var advisories []stateDeptAdvisory
	if err := json.NewDecoder(resp.Body).Decode(&advisories); err != nil {
		fr.Err = fmt.Errorf("safety decode: %w", err)
		return fr
	}

	// Match by country name prefix in title (e.g. "France - Level 2: ...")
	prefix := countryName + " - "
	var match *stateDeptAdvisory
	for i := range advisories {
		if strings.HasPrefix(advisories[i].Title, prefix) {
			match = &advisories[i]
			break
		}
	}
	if match == nil {
		fr.Err = fmt.Errorf("safety: no advisory data for %q", countryName)
		return fr
	}

	sd := SafetyData{
		Score:   parseLevelScore(match.Title),
		Sources: 1,
		Message: stripHTML(match.Summary),
		Updated: match.Updated,
	}

	data, err := json.Marshal(sd)
	if err != nil {
		fr.Err = fmt.Errorf("safety marshal: %w", err)
		return fr
	}

	fr.Data = data
	fr.FetchedAt = time.Now()
	return fr
}

// parseLevelScore extracts the advisory level from a title like
// "France - Level 2: Exercise Increased Caution" and maps it to a score.
func parseLevelScore(title string) float64 {
	// Level 1 → 1.0, Level 2 → 2.5, Level 3 → 3.5, Level 4 → 4.5
	scores := map[string]float64{
		"Level 1": 1.0,
		"Level 2": 2.5,
		"Level 3": 3.5,
		"Level 4": 4.5,
	}
	for k, v := range scores {
		if strings.Contains(title, k) {
			return v
		}
	}
	return 0
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// FetchCities retrieves all capital cities from the REST Countries API.
func FetchCities(ctx context.Context, client *http.Client, baseURL string) ([]string, error) {
	url := fmt.Sprintf("%s/v3.1/all?fields=capital", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cities request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cities call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cities: unexpected status %d", resp.StatusCode)
	}

	var items []struct {
		Capital []string `json:"capital"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("cities decode: %w", err)
	}

	cities := make([]string, 0, len(items))
	for _, item := range items {
		if len(item.Capital) > 0 && item.Capital[0] != "" {
			cities = append(cities, strings.ToLower(item.Capital[0]))
		}
	}
	return cities, nil
}
