package destination

import (
	"encoding/json"
	"time"
)

// FeedResult is the common intermediate type produced by all API fetchers.
type FeedResult struct {
	Source    string          // "weather", "country", "safety"
	City      string
	Data      json.RawMessage
	FetchedAt time.Time
	Err       error
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
