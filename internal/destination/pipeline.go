package destination

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	geocodeBaseURL       = "https://geocoding-api.open-meteo.com"
	openMeteoBaseURL     = "https://api.open-meteo.com"
	restCountriesBaseURL = "https://restcountries.com"
	advisoryBaseURL      = "https://www.travel-advisory.info/api"
)

// Pipeline orchestrates periodic polling of external APIs and batched upserts.
type Pipeline struct {
	repo     Repository
	client   *http.Client
	ch       chan FeedResult
	interval time.Duration
	debounce time.Duration
	geoCache sync.Map // city -> *GeocodingResult
}

// NewPipeline creates a Pipeline. If client is nil a default with 10s timeout is used.
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

// Start launches the listener and polls all cities on a ticker.
// It blocks until ctx is cancelled.
func (p *Pipeline) Start(ctx context.Context, cities []string) {
	listenerDone := make(chan struct{})
	go func() {
		p.listen(ctx)
		close(listenerDone)
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
			<-listenerDone
			return
		}
	}
}

// pollAll launches a goroutine per city, waits for all to finish.
func (p *Pipeline) pollAll(ctx context.Context, cities []string) {
	var wg sync.WaitGroup
	for _, city := range cities {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			p.pollCity(ctx, c)
		}(city)
	}
	wg.Wait()
}

// pollCity fetches geocoding (cached), weather, country, and safety data for a city.
func (p *Pipeline) pollCity(ctx context.Context, city string) {
	// Resolve geocoding (cached)
	var geo *GeocodingResult
	if v, ok := p.geoCache.Load(city); ok {
		geo = v.(*GeocodingResult)
	} else {
		var err error
		geo, err = fetchGeocode(ctx, p.client, geocodeBaseURL, city)
		if err != nil {
			log.Printf("pipeline: geocode %s: %v", city, err)
			return
		}
		p.geoCache.Store(city, geo)
	}

	// Weather (send immediately)
	wr := fetchWeather(ctx, p.client, openMeteoBaseURL, city, geo.Latitude, geo.Longitude)
	p.ch <- wr

	// Country and safety in parallel
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		cr := fetchCountry(ctx, p.client, restCountriesBaseURL, geo.CountryCode)
		cr.City = city
		p.ch <- cr
	}()
	go func() {
		defer wg.Done()
		sr := fetchSafety(ctx, p.client, advisoryBaseURL, city, geo.CountryCode)
		p.ch <- sr
	}()
	wg.Wait()
}

// listen collects FeedResults and flushes them in batches after a debounce period.
func (p *Pipeline) listen(ctx context.Context) {
	batch := make(map[string][]FeedResult)
	timer := time.NewTimer(p.debounce)
	timer.Stop() // start stopped — we only fire after receiving data

	for {
		select {
		case fr := <-p.ch:
			if fr.Err != nil {
				log.Printf("pipeline: skipping errored %s result for %s: %v", fr.Source, fr.City, fr.Err)
				continue
			}
			batch[fr.City] = append(batch[fr.City], fr)
			// Reset the debounce timer
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(p.debounce)

		case <-timer.C:
			p.flush(ctx, batch)
			batch = make(map[string][]FeedResult)

		case <-ctx.Done():
			// Drain any remaining items from channel
			draining := true
			for draining {
				select {
				case fr := <-p.ch:
					if fr.Err == nil {
						batch[fr.City] = append(batch[fr.City], fr)
					}
				default:
					draining = false
				}
			}
			if len(batch) > 0 {
				p.flush(context.Background(), batch)
			}
			return
		}
	}
}

// flush merges all FeedResults per city and upserts each destination.
func (p *Pipeline) flush(ctx context.Context, batch map[string][]FeedResult) {
	for city, results := range batch {
		merged := make(map[string]json.RawMessage, len(results))
		for _, fr := range results {
			merged[fr.Source] = fr.Data
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

		// Enrich with geo data if available
		if v, ok := p.geoCache.Load(city); ok {
			geo := v.(*GeocodingResult)
			d.Latitude = geo.Latitude
			d.Longitude = geo.Longitude
			d.Country = geo.Country
		}

		if err := p.repo.Upsert(ctx, d); err != nil {
			log.Printf("pipeline: upsert %s: %v", city, err)
		}
	}
}
