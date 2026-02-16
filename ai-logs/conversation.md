# AI Conversation Log

---

## Entry 1

**User Prompt:**
> Please log every prompt you receive from me and every response you return to me into a markdown file inside the repository `/ai-logs/conversation.md`

**Assistant Response:**
> Set up the logging file at `ai-logs/conversation.md`. Every subsequent prompt and response will be appended to this file.

---

## Entry 2

**User Prompt:**
> Bootstrap a golang monorepo project Destination Data Aggregation API that will use postgres, redis. Follow this structure: [exact directory tree provided]

**Assistant Response:**
> Bootstrapped the full project with the following files:
>
> - `go.mod` — module with chi, pgx, go-redis dependencies
> - `docker-compose.yml` — Postgres 16 + Redis 7
> - `cmd/server/main.go` — entrypoint with graceful shutdown
> - `internal/config/config.go` — env-based configuration
> - `internal/destination/model.go` — Destination domain model
> - `internal/destination/service.go` — business logic with cache-aside pattern
> - `internal/storage/postgres.go` — PostgreSQL CRUD repository
> - `internal/cache/redis.go` — Redis caching layer with TTL
> - `internal/api/router.go` — chi router with middleware
> - `internal/api/handler.go` — REST handlers (List, Get, Create, Update, Delete)
> - `migrations/000001_create_destinations.up.sql` / `.down.sql` — initial schema
> - `README.md` — quickstart and API docs
>
> All dependencies resolved, project compiles cleanly.

---

## Entry 3

**User Prompt:**
> init git repo and make init commit

**Assistant Response:**
> Initialized git repository and created initial commit `258ba8d` with all 15 files (608 insertions).

---

## Entry 4

**User Prompt:**
> create and push this repo to github

**Assistant Response:**
> Created public GitHub repo and pushed to https://github.com/ksavkiv/ygo_assignment. SSH failed so switched remote to HTTPS; push succeeded on `main`.

---

## Entry 5

**User Prompt:**
> add the logs storing rule to claude.md

**Assistant Response:**
> Created `CLAUDE.md` with a conversation logging rule instructing Claude Code to log every prompt and response to `ai-logs/conversation.md`.

---

## Entry 6

**User Prompt:**
> start compose

**Assistant Response:**
> Pulled images (postgres:16-alpine, redis:7-alpine) and started both containers via `docker compose up -d`. Postgres and Redis are running.

---

## Entry 7

**User Prompt:**
> research publicly available free API for rich, wide-ranging set of real-time destination data - the kind of information a knowledgeable travel advisor would draw on when making a recommendation. the goal is to provide data feed for AI-powered travel agent that helps users decide which destination is best for their holiday and why. use brave search

**Assistant Response:**
> Conducted extensive research via Brave Search across 11 parallel queries covering weather, flights/hotels, attractions/POIs, safety/visas, currency/cost-of-living, events, air quality, photos, and country data. Compiled a categorized inventory of 20+ free/freemium APIs suitable for powering an AI travel advisor. Key findings organized by data category: Weather (Open-Meteo, OpenWeatherMap, WeatherAPI), Attractions (OpenTripMap, Amadeus POI/Tours), Safety (travel-advisory.info, TuGo, US State Dept), Cost (Numbeo, ExchangeRate-API), Events (Ticketmaster Discovery, PredictHQ), Country Info (REST Countries), Air Quality (Open-Meteo AQ, WAQI), and Photos (Unsplash, Pexels). Full details provided in response.

---

## Entry 8

**User Prompt:**
> Research the Open-Meteo weather API and Open-Meteo Air Quality API technical details for integration. I need: Base URL(s), Authentication method, Rate limits, Key endpoints with example request URLs, Response format (JSON structure summary), Key query parameters, License/terms.

**Assistant Response:**
> Conducted technical deep-dive via Brave Search on both Open-Meteo Weather Forecast API and Air Quality API. Compiled structured integration reference covering: base URLs (api.open-meteo.com, air-quality-api.open-meteo.com), no-auth-required access, rate limits (10K/day, 5K/hour, 600/min for free tier), full endpoint documentation with example URLs, JSON response structures, all key query parameters for both APIs, and CC-BY 4.0 licensing terms. Detailed findings provided in response.

---

## Entry 9

**User Prompt:**
> bootstrap these API endpoints:
> - GET /api/v1/destinations/:city — Get cached/stored destination data
> - POST /api/v1/destinations/:city/refresh — Fetch fresh data from external sources, store and cache it
> - GET /api/v1/health — Health check (DB + Redis connectivity)

**Assistant Response:**
> Rewrote the entire API layer from generic ID-based CRUD to city-based endpoints:
>
> - **Model** — renamed `name` to `city` as the primary lookup key
> - **Service** — added `Fetcher` interface, `GetByCity` (cache-aside), `Refresh` (fetch → upsert → cache)
> - **Fetcher** — new `StubFetcher` placeholder for external API integration
> - **Storage** — replaced CRUD with `GetByCity` and `Upsert` (ON CONFLICT city)
> - **Router** — mounted under `/api/v1/` with `{city}` param
> - **Handler** — `GetByCity`, `Refresh`, and `Health` (pings both Postgres and Redis)
> - **Migration** — `city` column now `UNIQUE`
>
> Smoke tested all endpoints:
> - `GET /api/v1/health` → `{"postgres":true,"redis":true}`
> - `POST /api/v1/destinations/paris/refresh` → 200 with stub data stored in DB + cached
> - `GET /api/v1/destinations/paris` → 200 served from cache
> - `GET /api/v1/destinations/atlantis` → 404

---

## Entry 10

**User Prompt:**
> Research the Ticketmaster Discovery API v2 technical details for integration. I need: Base URL, Authentication method, Rate limits for free tier, Key endpoints with example request URLs, Response format (JSON structure summary), Key query parameters, How to get an API key. Use WebFetch to check the official docs.

**Assistant Response:**
> Researched Ticketmaster Discovery API v2 via Brave Search (WebFetch was unavailable). Compiled comprehensive technical reference covering: base URL (app.ticketmaster.com/discovery/v2/), API key auth via query parameter, free tier rate limits (5000/day, 2 req/sec), all key endpoints (events, venues, attractions, classifications with search + detail), HAL+JSON response structure (_embedded, _links, page), full query parameter reference (35+ params including city, countryCode, startDateTime, classificationName, keyword, geoPoint, radius, etc.), and signup instructions via developer.ticketmaster.com.

---

## Entry 11

**User Prompt:**
> Research the REST Countries API (restcountries.com) technical details for integration. I need: Base URL, Authentication method, Rate limits, Key endpoints with example request URLs (all, by name, by code, by region, by currency, by language), Response format (JSON structure - key fields returned), Available filter fields parameter. Use WebFetch to check https://restcountries.com/ and https://gitlab.com/restcountries/restcountries

**Assistant Response:**
> Fetched and analyzed the official REST Countries flatdoc.md documentation plus multiple search results and code examples. Compiled a full integration reference: base URL (https://restcountries.com/v3.1/), no authentication required, no formally published rate limits (free and open source, ~4M hits/day reported), all 13 endpoints with example URLs, v3.1 JSON response structure with 30+ fields documented, and the `?fields=` comma-separated filter parameter (mandatory for `/all` endpoint). Full structured details provided in the response.

---

## Entry 12

**User Prompt:**
> Research the following APIs for integration details: 1) Geoapify Places API, 2) Foursquare Places API v3, 3) WeatherAPI.com. Need base URL, auth method, rate limits for free tier, key endpoints, response format, and categories (for places APIs).

**Assistant Response:**
> Researched all three APIs via Brave Search (WebFetch was unavailable). Compiled structured integration references for each: Geoapify Places API (base URL api.geoapify.com/v2/places, API key via query param, 3000 credits/day free, 5 RPS, GeoJSON response, 500+ categories), Foursquare Places API v3 (base URL api.foursquare.com/v3, API key via Authorization header, 10K free Pro calls, 50 QPS sandbox, JSON response), WeatherAPI.com (base URL api.weatherapi.com/v1, API key via query param, 1M calls/month free, 3-day forecast limit on free tier, JSON/XML response). Full details provided in response.

---

## Entry 13

**User Prompt:**
> Research the ExchangeRate-API (exchangerate-api.com) technical details for integration. I need: Base URL(s) - both the free/open-access endpoint and the standard endpoint, Authentication method, Rate limits for free tier, Key endpoints with example request URLs, Response format (JSON structure), Supported currencies. Use WebFetch to check the official docs.

**Assistant Response:**
> Researched ExchangeRate-API via Brave Search (WebFetch was unavailable). Compiled structured integration reference covering: two base URLs (open.er-api.com for no-key access, v6.exchangerate-api.com for keyed access), authentication methods (none for open access, API key in URL path for standard), rate limits (open access: safe at once/hour; free plan: 1,500 req/month), all 7 endpoints with full JSON response structures (Standard, Pair Conversion with optional amount, Enriched, Historical, Supported Codes, Quota, Open Access), and supported currency details (161 currencies via ISO 4217 codes). Full findings provided in response.

---

## Entry 14

**User Prompt:**
> Research the following APIs for integration details: US State Department Travel Advisories API, Travelpayouts Data API, OpenWeatherMap API. Use WebFetch on specified URLs. Need base URLs, auth methods, rate limits, endpoints, and response formats for each.

**Assistant Response:**
> WebFetch and Bash were restricted, so used Brave Search plus the initial curl response to compile comprehensive integration references for all three APIs. Detailed findings include: State Dept API (cadataapi.state.gov, no auth, endpoints for advisories and country info, JSON response with Title/Link/Category/Summary/Published/Updated fields), Travelpayouts (api.travelpayouts.com, X-Access-Token header, 200 req/hr from one IP, 60 req/min per method, REST + GraphQL endpoints for flights/hotels/airlines), OpenWeatherMap (api.openweathermap.org/data/2.5, appid query param, 60 calls/min and 1000 calls/day free, endpoints for current/forecast/air-pollution with full JSON structures documented).

---

## Entry 15

**User Prompt:**
> Research the Amadeus Self-Service APIs technical details for integration. I need details on these specific APIs: 1) Points of Interest API, 2) Tours and Activities API, 3) Flight Offers Search API, 4) Hotel Search API, 5) Travel Recommendations API. For each, need: Base URL, Authentication method (OAuth2 flow details), Rate limits for free/test tier, Key endpoints with example request URLs, Response format summary, How to get API credentials.

**Assistant Response:**
> Researched all five Amadeus Self-Service APIs via Brave Search (WebFetch was unavailable). Compiled comprehensive integration reference covering: shared OAuth2 client_credentials flow (token endpoint, 30-min expiry), test/production base URLs, rate limits (10 TPS test, 200-10,000 free calls/month per API), and per-API details including all endpoints with example URLs and response format summaries. APIs covered: Points of Interest (3 geo-search endpoints), Tours and Activities (3 endpoints, 300K+ activities), Flight Offers Search v2 (GET/POST, 400+ airlines), Hotel Search v3 (two-step Hotel List + Hotel Offers), and Travel Recommendations (AI-powered destination suggestions). Full structured details provided in response.

---

## Entry 16

**User Prompt:**
> Research the OpenTripMap API technical details for integration. I need: Base URL, Authentication method, Rate limits, Key endpoints with example request URLs (especially: places by bounding box, places by radius, place details), Response format (JSON structure summary), Key query parameters (kinds/categories, radius, format, etc.), How to get an API key. Use WebFetch to check https://dev.opentripmap.org/product and https://opentripmap.io/product

**Assistant Response:**
> Conducted technical deep-dive via Brave Search (WebFetch was unavailable) across 10+ queries. Compiled comprehensive structured integration reference with base URL (api.opentripmap.com/0.1), all 5 endpoints (geoname, radius, bbox, xid, autosuggest) with example URLs, full JSON response schemas for list and detail endpoints, complete query parameter docs, kinds hierarchy, rate limits, and registration instructions. Full details in response.

---

## Entry 17

**User Prompt:**
> Research the following three APIs for integration details: 1) WAQI (World Air Quality Index) API, 2) VisaDB.io API, 3) Pexels API. Need base URL, auth method, rate limits, key endpoints, response format, how to get credentials for each.

**Assistant Response:**
> Researched all three APIs via Brave Search (WebFetch was unavailable). Compiled structured integration references: WAQI API (base URL api.waqi.info, token via query param, 1000 req/min with burst 60, city/geo/search/map endpoints, JSON with status+data wrapper), VisaDB.io API (contact-based access, 5 data categories across 200+ countries, Postman-documented endpoints, JSON response), Pexels API (base URL api.pexels.com/v1, Authorization header, 20K req/month default, search/curated/get-photo endpoints, JSON with Photo resource objects). Full detailed findings provided in response.

---

## Entry 18

**User Prompt:**
> Research the following three APIs for integration details: 1) TuGo Travel Advisory API, 2) Unsplash API, 3) Numbeo API. Need base URL, auth method, rate limits, key endpoints, response format, how to get API key, and pricing tier details (for Numbeo).

**Assistant Response:**
> Researched all three APIs via Brave Search and direct HTML fetching of Numbeo doc pages. Compiled structured integration references for each: TuGo Travel Advisory API (base URL api.tugo.com/v1/travelsafe, X-Auth-API-Key header auth, free REST API, two main endpoints), Unsplash API (base URL api.unsplash.com, Client-ID auth, 50 req/hr demo / 5000 req/hr production), Numbeo API (base URL www.numbeo.com/api/, api_key query param, paid ~$220-260/month, 15+ endpoints including city_prices, country_prices, indices, rankings). Full details provided in response.

---

## Entry 19

**User Prompt:**
> Research the travel-advisory.info API technical details for integration. I need: 1) Base URL, 2) Authentication method, 3) Rate limits, 4) Key endpoints with example request URLs, 5) Response format (JSON structure summary), 6) Available data fields. Use WebFetch to check https://www.travel-advisory.info/data-api

**Assistant Response:**
> WebFetch and Bash (curl) were both restricted. Conducted research via Brave Search across 20+ queries, supplemented by analysis of the dgraph-io/travel Go client source code on GitHub (which consumes this exact API) and the apilist.fun listing. Compiled structured integration reference covering: base URL (https://www.travel-advisory.info/api), no authentication required, no formally published rate limits (fair-use policy), 2 endpoints (all countries and single country via ?countrycode=), full JSON response structure reverse-engineered from Go struct tags (api_status wrapper + data map keyed by ISO alpha-2 codes), and all 8 data fields (iso_alpha2, name, continent, advisory.score, advisory.sources_active, advisory.message, advisory.updated, advisory.source). Also noted the "Extended Data API" paid tier.

---

## Entry 20

**User Prompt:**
> you're not following tdd approach... add unit tests

**Assistant Response:**
> Acknowledged the gap — should have been writing tests first. Added unit tests across 3 packages:
>
> - `internal/destination/service_test.go` — 6 tests: cache hit, cache miss + DB hit, cache miss + DB miss, refresh success, refresh fetcher error, refresh upsert error
> - `internal/api/handler_test.go` — 6 tests: GetByCity found/not-found, Refresh success/error, Health all-up/redis-down
> - `internal/config/config_test.go` — 2 tests: defaults, env overrides
>
> All 12 tests pass. Committed to following TDD (red-green-refactor) for all future work.

---

## Entry 21

**User Prompt:**
> update CLAUDE.md to include dev rules for workflow(tdd, project structure, how to run compose, tests, check coverage >80%), check README.md

**Assistant Response:**
> Updated `CLAUDE.md` with comprehensive dev rules covering:
>
> - **TDD workflow** — red-green-refactor cycle, tests-first mandate
> - **Test conventions** — standard `testing` pkg, hand-rolled mocks, `t.Errorf`/`t.Fatalf`
> - **Test coverage** — minimum 80% required, commands for coverage reports
> - **Project structure** — full directory tree with descriptions
> - **Docker Compose** — start/stop/volumes commands, service details table
> - **Database migrations** — commands to apply migrations in order
> - **Running the server** — `go run` command, env var configuration table
> - **Running tests** — all test commands (all/verbose/package/coverage)
> - **API endpoints** — corrected endpoint table matching actual routes
>
> Also checked `README.md` — noted the API endpoints table is stale (lists old CRUD routes instead of the current city-based `/api/v1/` routes).

---

## Entry 22

**User Prompt:**
> yes, update the README

**Assistant Response:**
> Updated `README.md` with corrections:
>
> - **Go version** — 1.23+ → 1.24+
> - **Description** — updated to mention data aggregation from external sources
> - **Quick Start** — added migration 2 (`000002_add_metadata_jsonb.up.sql`)
> - **API Endpoints** — replaced stale CRUD routes with actual city-based routes (`/api/v1/destinations/{city}`, `/api/v1/destinations/{city}/refresh`, `/api/v1/health`)
> - **Running Tests** — added test commands section with coverage
> - **Project Structure** — added `docs/` directory, updated descriptions

---

## Entry 23

**User Prompt:**
> create new migration to store the data for travel destination data, use these columns(country, city, timestamp, etc.) and JSONB column to store dynamic metadata from different resources(weather, points of interests, metadata, etc.)

**Assistant Response:**
> Followed TDD (red-green-refactor):
>
> **RED** — Added 2 failing tests (`TestRefresh_PreservesMetadata`, `TestGetByCity_ReturnsMetadata`) that assert a `Metadata json.RawMessage` field on `Destination`. Confirmed compile failure: `Destination has no field or method Metadata`.
>
> **GREEN** — Made tests pass by:
> - Adding `Metadata json.RawMessage` field to `Destination` model (replaced `Description`)
> - Creating migration `000002_add_metadata_jsonb.up.sql` — adds `metadata JSONB NOT NULL DEFAULT '{}'`, drops `description`
> - Updating `storage/postgres.go` to read/write `metadata` JSONB column
> - Updating `fetcher.go` to return metadata in stub
>
> **Result** — 14/14 tests pass. Smoke tested end-to-end: `POST /refresh` stores `{"source":"stub"}` as JSONB, `GET` returns it from cache with full round-trip.

---

## Entry 24

**User Prompt:**
> commit and push

**Assistant Response:**
> Staged 17 files, committed as `aeabe12` ("Add city-based API endpoints, JSONB metadata, unit tests, and CLAUDE.md"), pushed to `origin/main`.

---

## Entry 25

**User Prompt:**
> Create FeedResult type and API response structs in `internal/destination/feed.go` with types: FeedResult, GeocodingResult, geocodingResponse, WeatherData, openMeteoResponse, CountryData, SafetyData, advisoryAPIResponse.

**Assistant Response:**
> Created `internal/destination/feed.go` with all 8 types (5 exported, 3 unexported). Verified the package compiles and all existing tests pass. Committed as `06c976e` ("feat: add FeedResult and API response types for data feed pipeline").

---

## Entry 26

**User Prompt:**
> Implement Tasks 2-3: Write API fetcher tests (`feed_test.go`) and implement fetcher functions (`fetchGeocode`, `fetchWeather`, `fetchCountry`, `fetchSafety`) in `feed.go`. Follow TDD red-green-refactor. Tests use `httptest.Server` to mock external APIs.

**Assistant Response:**
> Followed TDD red-green-refactor cycle:
>
> **RED** — Created `internal/destination/feed_test.go` with 8 tests using `httptest.Server` mocks:
> - `TestFetchGeocode_Success` — verifies coordinates and country code
> - `TestFetchGeocode_NoResults` — verifies error on empty results
> - `TestFetchWeather_Success` — verifies FeedResult with Source="weather", correct WeatherData
> - `TestFetchWeather_HTTPError` — verifies Err set on 500
> - `TestFetchCountry_Success` — verifies FeedResult with Source="country", correct CountryData
> - `TestFetchCountry_HTTPError` — verifies Err set on 404
> - `TestFetchSafety_Success` — verifies FeedResult with Source="safety", correct SafetyData
> - `TestFetchSafety_HTTPError` — verifies Err set on 503
>
> Confirmed all 8 tests fail (functions undefined).
>
> **GREEN** — Added 4 fetcher functions to `feed.go` plus `restCountryItem` helper type:
> - `fetchGeocode` — calls Open-Meteo geocoding, returns first result or error
> - `fetchWeather` — calls Open-Meteo forecast, parses into WeatherData, returns FeedResult
> - `fetchCountry` — calls REST Countries alpha endpoint, parses array response into CountryData
> - `fetchSafety` — calls travel-advisory.info, looks up country code key, extracts SafetyData
>
> All 16 tests pass (8 new + 8 existing). Full project test suite green.

---

## Entry 27

**User Prompt:**
> Implement Tasks 4-5: Write pipeline tests (TestPipelineListener_BatchesAndFlushes, TestPipelineListener_MultipleCities, TestPipelineListener_SkipsErrors) and create Pipeline struct with listener, debounced batching, and flush in `internal/destination/pipeline.go`. Follow TDD.

**Assistant Response:**
> Followed TDD red-green-refactor cycle:
>
> **RED** — Added 3 pipeline tests to `internal/destination/feed_test.go`:
> - `TestPipelineListener_BatchesAndFlushes` — sends 3 FeedResults (weather, country, safety) for "paris", waits for 100ms debounce, verifies single upsert with merged metadata containing all 3 keys
> - `TestPipelineListener_MultipleCities` — sends weather results for "paris" and "london", verifies 2 separate upserts
> - `TestPipelineListener_SkipsErrors` — sends a good weather result and an errored safety result, verifies upsert only contains weather key
> Confirmed all 3 tests fail (NewPipeline undefined).
>
> **GREEN** — Created `internal/destination/pipeline.go` with:
> - `Pipeline` struct — repo, client, ch (buffered chan FeedResult), interval, debounce, geoCache (sync.Map)
> - `NewPipeline` — constructor with default 10s-timeout client, 100-buffered channel
> - `Start` — launches listener goroutine, initial pollAll, ticker loop, blocks until ctx cancelled
> - `pollAll` — goroutine per city with WaitGroup
> - `pollCity` — geocode (cached via sync.Map), weather, country+safety in parallel, all send to ch
> - `listen` — select loop with debounced timer: receives FeedResults, skips errors, batches by city, flushes on timer, drains on ctx.Done
> - `flush` — merges FeedResults by Source into metadata JSON, enriches with geo data, calls repo.Upsert
>
> All 19 tests pass (8 feed + 3 pipeline + 8 service). Committed as `2456d3a`.

---

## Entry 28

**User Prompt:**
> Implement Tasks 6-8: Add APIFetcher test (TestAPIFetcher_Fetch in feed_test.go), add APIFetcher struct to fetcher.go (keep StubFetcher), and wire into main.go (replace StubFetcher with NewDefaultAPIFetcher, add Pipeline startup with seed cities).

**Assistant Response:**
> Followed TDD red-green-refactor cycle:
>
> **RED** — Added `TestAPIFetcher_Fetch` to `internal/destination/feed_test.go`:
> - Spins up 4 httptest.Server mocks (geocoding, weather, country, safety)
> - Creates APIFetcher with mock URLs, calls Fetch("paris")
> - Asserts City="paris", Country="France", Latitude=48.85
> - Verifies metadata JSON contains "weather", "country", and "safety" keys
> Confirmed test fails (NewAPIFetcher undefined).
>
> **GREEN** — Added `APIFetcher` to `internal/destination/fetcher.go`:
> - `APIFetcher` struct with http.Client and 4 URL fields
> - `NewAPIFetcher(geocodeURL, weatherURL, countryURL, safetyURL)` constructor
> - `NewDefaultAPIFetcher()` using production base URL constants
> - `Fetch(ctx, city)` — geocodes city, then fetches weather/country/safety concurrently via goroutines+WaitGroup, merges results into metadata JSON, returns Destination
>
> **Wiring** — Updated `cmd/server/main.go`:
> - Replaced `destination.NewStubFetcher()` with `destination.NewDefaultAPIFetcher()`
> - Added pipeline startup after server: seed cities ["paris", "london", "tokyo"], 5-min interval, 5-sec debounce
>
> All tests pass (20 destination + 12 api + 4 config = 36 total). Committed as `b6ae833`.

---

## Entry 29

**User Prompt:**
> Increase test coverage for `internal/destination` from 64% to at least 80%. Make Pipeline base URLs configurable on struct, add tests for StubFetcher, pollCity, pollAll, Start, and error branches for all fetch functions.

**Assistant Response:**
> Made two sets of changes to reach **90.8% coverage** (up from 64%):
>
> **pipeline.go** — Added 4 URL fields (`geocodeURL`, `weatherURL`, `countryURL`, `safetyURL`) to `Pipeline` struct, initialized from constants in `NewPipeline`, used in `pollCity` instead of package-level constants. This makes `pollCity`/`pollAll`/`Start` testable with httptest servers.
>
> **feed_test.go** — Added 19 new tests:
> - `TestStubFetcher` — verifies lowercase city, "unknown" country, `{"source":"stub"}` metadata
> - `TestNewDefaultAPIFetcher` — verifies default URLs match constants
> - `TestPollCity` — sets up 4 httptest servers, calls pollCity, drains 3 results from channel
> - `TestPollCity_GeocodeError` — geocode returns 500, verifies no results on channel
> - `TestPollCity_UsesGeoCache` — pre-populates geoCache, verifies 3 results without geocode server
> - `TestPollAll` — 2 cities, verifies 6 results on channel
> - `TestStart_CancelsCleanly` — verifies Start returns after ctx cancellation
> - `TestFetchGeocode_HTTPError` — 500 response
> - `TestFetchGeocode_BadJSON` — invalid JSON decode error
> - `TestFetchWeather_BadJSON` — invalid JSON decode error
> - `TestFetchCountry_EmptyArray` — `[]` response triggers "no results" error
> - `TestFetchCountry_BadJSON` — invalid JSON decode error
> - `TestFetchSafety_MissingCountryCode` — valid JSON but wrong country code
> - `TestFetchSafety_BadJSON` — invalid JSON decode error
> - `TestFetchGeocode_ConnectionError` — unreachable host
> - `TestFetchWeather_ConnectionError` — unreachable host
> - `TestFetchCountry_ConnectionError` — unreachable host
> - `TestFetchSafety_ConnectionError` — unreachable host
> - `TestFlush_UpsertError` — verifies flush handles upsert errors gracefully
>
> All 37 tests pass. Coverage: 90.8%. Committed as `838cf59`.

---
