# External API Resources — Destination Data Aggregation

Technical reference for all external APIs used to power the AI travel advisor data feed.

---

## Table of Contents

1. [Weather & Climate](#1-weather--climate)
2. [Attractions & Points of Interest](#2-attractions--points-of-interest)
3. [Travel Safety & Advisories](#3-travel-safety--advisories)
4. [Cost of Living & Currency](#4-cost-of-living--currency)
5. [Flights & Travel Costs](#5-flights--travel-costs)
6. [Events & Festivals](#6-events--festivals)
7. [Country & Destination Metadata](#7-country--destination-metadata)
8. [Air Quality & Environmental](#8-air-quality--environmental)
9. [Destination Photos](#9-destination-photos)

---

## 1. Weather & Climate

### 1.1 Open-Meteo (Weather Forecast)

| Field | Value |
|---|---|
| **Base URL** | `https://api.open-meteo.com/v1/forecast` |
| **Auth** | None required |
| **Rate Limits** | 10,000 req/day, 5,000 req/hour, 600 req/min (free non-commercial) |
| **Response Format** | JSON |
| **License** | CC-BY 4.0 (attribution required) |
| **Docs** | https://open-meteo.com/en/docs |

**Key Endpoints:**

```
GET /v1/forecast?latitude=52.52&longitude=13.41&daily=temperature_2m_max,temperature_2m_min,precipitation_sum,weathercode&timezone=Europe/Berlin
```

**Key Query Parameters:**

| Parameter | Description |
|---|---|
| `latitude`, `longitude` | Required. Location coordinates |
| `hourly` | Comma-separated: `temperature_2m`, `relative_humidity_2m`, `precipitation`, `weathercode`, `windspeed_10m`, `uv_index` |
| `daily` | Comma-separated: `temperature_2m_max`, `temperature_2m_min`, `precipitation_sum`, `weathercode`, `sunrise`, `sunset`, `uv_index_max` |
| `current_weather` | `true` to include current conditions |
| `timezone` | IANA timezone string (e.g. `Europe/Berlin`, `auto`) |
| `forecast_days` | 1–16 (default 7) |
| `past_days` | 0–92 for historical data |
| `temperature_unit` | `celsius` (default) or `fahrenheit` |

**Response Structure:**
```json
{
  "latitude": 52.52,
  "longitude": 13.41,
  "generationtime_ms": 0.5,
  "utc_offset_seconds": 3600,
  "timezone": "Europe/Berlin",
  "current_weather": {
    "temperature": 15.3,
    "windspeed": 12.1,
    "winddirection": 210,
    "weathercode": 3,
    "time": "2026-02-16T14:00"
  },
  "daily": {
    "time": ["2026-02-16", "2026-02-17"],
    "temperature_2m_max": [16.2, 14.8],
    "temperature_2m_min": [8.1, 7.3],
    "precipitation_sum": [0.0, 2.5],
    "weathercode": [3, 61]
  }
}
```

**Geocoding Helper** (resolve city name to coordinates):
```
GET https://geocoding-api.open-meteo.com/v1/search?name=Paris&count=1
```

---

### 1.2 OpenWeatherMap

| Field | Value |
|---|---|
| **Base URL** | `https://api.openweathermap.org/data/2.5` |
| **Auth** | API key via `appid` query parameter |
| **Rate Limits** | 60 calls/min, 1,000 calls/day (free tier) |
| **Response Format** | JSON |
| **Get API Key** | https://home.openweathermap.org/users/sign_up |

**Key Endpoints:**

```
# Current weather
GET /data/2.5/weather?q=London&appid={key}&units=metric

# 5-day / 3-hour forecast
GET /data/2.5/forecast?q=Paris&appid={key}&units=metric

# Air pollution
GET /data/2.5/air_pollution?lat=48.85&lon=2.35&appid={key}
```

**Response Structure (current weather):**
```json
{
  "coord": {"lon": -0.1257, "lat": 51.5085},
  "weather": [{"id": 800, "main": "Clear", "description": "clear sky", "icon": "01d"}],
  "main": {"temp": 15.3, "feels_like": 14.1, "humidity": 72, "pressure": 1013},
  "wind": {"speed": 4.1, "deg": 230},
  "clouds": {"all": 0},
  "dt": 1708099200,
  "sys": {"country": "GB", "sunrise": 1708067100, "sunset": 1708102500},
  "name": "London"
}
```

---

### 1.3 WeatherAPI.com

| Field | Value |
|---|---|
| **Base URL** | `https://api.weatherapi.com/v1` |
| **Auth** | API key via `key` query parameter |
| **Rate Limits** | 1,000,000 calls/month (free tier) |
| **Free Tier Limits** | 3-day forecast max, no historical data |
| **Response Format** | JSON or XML |
| **Get API Key** | https://www.weatherapi.com/signup.aspx |

**Key Endpoints:**

```
# Current + forecast
GET /v1/forecast.json?key={key}&q=Paris&days=3&aqi=yes

# Search/autocomplete
GET /v1/search.json?key={key}&q=Lon

# Astronomy (sunrise/sunset/moon)
GET /v1/astronomy.json?key={key}&q=Paris&dt=2026-02-16
```

---

## 2. Attractions & Points of Interest

### 2.1 OpenTripMap

| Field | Value |
|---|---|
| **Base URL** | `https://api.opentripmap.com/0.1` |
| **Auth** | API key via `apikey` query parameter |
| **Rate Limits** | 5 req/sec (free tier) |
| **Coverage** | 10M+ tourist attractions worldwide |
| **Data Sources** | OpenStreetMap, Wikidata, Wikipedia |
| **Response Format** | JSON or GeoJSON |
| **Get API Key** | https://opentripmap.io/product (free registration) |

**Key Endpoints:**

```
# Resolve city name to geoname
GET /0.1/en/places/geoname?name=Paris&apikey={key}

# Search by radius
GET /0.1/en/places/radius?radius=5000&lon=2.3522&lat=48.8566&kinds=cultural,museums&rate=3&format=json&limit=20&apikey={key}

# Search by bounding box
GET /0.1/en/places/bbox?lon_min=2.25&lat_min=48.81&lon_max=2.42&lat_max=48.90&kinds=historic&format=json&apikey={key}

# Place details by XID
GET /0.1/en/places/xid/W28078508?apikey={key}

# Autosuggest
GET /0.1/en/places/autosuggest?name=Eiffel&radius=50000&lon=2.35&lat=48.85&format=json&apikey={key}
```

**Key Parameters:**

| Parameter | Description |
|---|---|
| `kinds` | Comma-separated categories: `interesting_places`, `amusements`, `sport`, `cultural`, `museums`, `historic`, `architecture`, `natural`, `religion`, `beaches`, `foods`, `shops`, `transport`, `banks`, `accommodations` |
| `rate` | Minimum popularity rating: `1` (any), `2` (interesting), `3` (must-see), `3h` (heritage) |
| `radius` | Search radius in meters (max 40,000) |
| `format` | `json`, `geojson`, or `count` |
| `limit` | Max results (default 500) |

**Response Structure (list):**
```json
[
  {
    "xid": "W28078508",
    "name": "Eiffel Tower",
    "rate": 7,
    "osm": "way/28078508",
    "wikidata": "Q243",
    "kinds": "architecture,towers,interesting_places",
    "point": {"lon": 2.2945, "lat": 48.8584}
  }
]
```

**Response Structure (detail by xid):**
```json
{
  "xid": "W28078508",
  "name": "Eiffel Tower",
  "address": {"city": "Paris", "road": "Avenue Anatole France", "state": "Ile-de-France", "country": "France", "postcode": "75007"},
  "rate": "7",
  "osm": "way/28078508",
  "wikidata": "Q243",
  "kinds": "architecture,towers,interesting_places",
  "sources": {"geometry": "osm", "attributes": ["osm", "wikidata", "wikipedia"]},
  "otm": "https://opentripmap.com/en/card/W28078508",
  "wikipedia": "https://en.wikipedia.org/wiki/Eiffel_Tower",
  "image": "https://commons.wikimedia.org/wiki/...",
  "preview": {"source": "https://upload.wikimedia.org/...", "height": 600, "width": 800},
  "wikipedia_extracts": {"title": "Eiffel Tower", "text": "The Eiffel Tower is a wrought-iron lattice tower..."},
  "point": {"lon": 2.2945, "lat": 48.8584}
}
```

---

### 2.2 Amadeus — Points of Interest

| Field | Value |
|---|---|
| **Base URL** | Test: `https://test.api.amadeus.com` / Prod: `https://api.amadeus.com` |
| **Auth** | OAuth2 `client_credentials` grant (see [Amadeus Auth](#amadeus-authentication)) |
| **Rate Limits** | 10 TPS (test), free production allowance then pay-per-call |
| **Response Format** | JSON |
| **Get Credentials** | https://developers.amadeus.com/register |

**Endpoints:**

```
# By geo-coordinates
GET /v1/reference-data/locations/pois?latitude=48.8566&longitude=2.3522&radius=2

# By bounding box
GET /v1/reference-data/locations/pois/by-square?north=48.90&west=2.25&south=48.81&east=2.42

# By ID
GET /v1/reference-data/locations/pois/{poisId}
```

---

### 2.3 Amadeus — Tours & Activities

```
# By geo-coordinates
GET /v1/shopping/activities?latitude=48.8566&longitude=2.3522&radius=5

# By bounding box
GET /v1/shopping/activities/by-square?north=48.90&west=2.25&south=48.81&east=2.42

# By ID
GET /v1/shopping/activities/{activityId}
```

Returns: id, name, shortDescription, geoCode, rating, price (amount + currencyCode), pictures, bookingLink, minimumDuration. 300,000+ activities worldwide.

---

### Amadeus Authentication

All Amadeus Self-Service APIs share the same OAuth2 flow:

```bash
curl -X POST "https://test.api.amadeus.com/v1/security/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id={API_KEY}&client_secret={API_SECRET}"
```

Response:
```json
{
  "type": "amadeusOAuth2Token",
  "username": "you@example.com",
  "application_name": "MyApp",
  "client_id": "abc123",
  "token_type": "Bearer",
  "access_token": "eyJhbGci...",
  "expires_in": 1799,
  "state": "approved"
}
```

Use token: `Authorization: Bearer {access_token}` — expires in ~30 minutes.

---

### 2.4 Geoapify Places

| Field | Value |
|---|---|
| **Base URL** | `https://api.geoapify.com/v2/places` |
| **Auth** | API key via `apiKey` query parameter |
| **Rate Limits** | 3,000 credits/day free, 5 requests/sec |
| **Response Format** | GeoJSON |
| **Get API Key** | https://myprojects.geoapify.com/ |
| **Categories** | 500+ (e.g., `tourism.sights`, `catering.restaurant`, `accommodation.hotel`) |

```
GET /v2/places?categories=tourism.sights&filter=circle:2.3522,48.8566,5000&limit=20&apiKey={key}
```

---

### 2.5 Foursquare Places API v3

| Field | Value |
|---|---|
| **Base URL** | `https://api.foursquare.com/v3` |
| **Auth** | `Authorization: {API_KEY}` header |
| **Rate Limits** | Free Starter plan available, 50 QPS sandbox |
| **Response Format** | JSON |
| **Get API Key** | https://foursquare.com/developers/signup |

```
# Place search
GET /v3/places/search?query=coffee&ll=48.8566,2.3522&radius=5000&categories=13065&limit=10

# Place details
GET /v3/places/{fsq_id}

# Place photos (Premium)
GET /v3/places/{fsq_id}/photos
```

---

## 3. Travel Safety & Advisories

### 3.1 travel-advisory.info

| Field | Value |
|---|---|
| **Base URL** | `https://www.travel-advisory.info/api` |
| **Auth** | None required |
| **Rate Limits** | No published limits (be respectful, cache results) |
| **Response Format** | JSON |
| **Docs** | https://www.travel-advisory.info/data-api |

**Key Endpoints:**

```
# All countries
GET https://www.travel-advisory.info/api

# Single country
GET https://www.travel-advisory.info/api?countrycode=AU
```

**Response Structure:**
```json
{
  "api_status": {
    "request": {"item": "AU"},
    "reply": {"cache": "cached", "code": 200, "status": "ok", "note": "..."}
  },
  "data": {
    "AU": {
      "iso_alpha2": "AU",
      "name": "Australia",
      "continent": "OC",
      "advisory": {
        "score": 2.7,
        "sources_active": 7,
        "message": "",
        "updated": "2026-02-15 07:28:22",
        "source": "https://www.travel-advisory.info/australia"
      }
    }
  }
}
```

Score ranges: 0–2.5 (low risk), 2.5–3.5 (medium), 3.5–4.5 (high), 4.5–5 (extreme).

---

### 3.2 US State Department Travel Advisories

| Field | Value |
|---|---|
| **Base URL** | `https://cadataapi.state.gov/api` |
| **Auth** | None required |
| **Rate Limits** | No published limits |
| **Response Format** | JSON (Atom feed format) |

**Endpoints:**

```
# All travel advisories
GET https://cadataapi.state.gov/api/TravelAdvisories
```

**Response Fields:** Title, Link, Category (Level 1–4), Summary (HTML), Published, Updated.

Advisory Levels: 1 (Exercise Normal Precautions), 2 (Exercise Increased Caution), 3 (Reconsider Travel), 4 (Do Not Travel).

---

### 3.3 TuGo Travel Advisory API

| Field | Value |
|---|---|
| **Base URL** | `https://api.tugo.com/v1/travelsafe` |
| **Auth** | `X-Auth-API-Key: {key}` header |
| **Rate Limits** | Not published (free API) |
| **Coverage** | 225+ countries |
| **Response Format** | JSON |
| **Get API Key** | https://developer.tugo.com (free registration) |

**Endpoints:**

```
# List all countries
GET /v1/travelsafe/countries

# Get country detail
GET /v1/travelsafe/countries/{countryCode}
```

Returns: advisories, health/safety info, climate/disaster updates, passport/entry requirements, embassy contacts.

---

### 3.4 VisaDB.io

| Field | Value |
|---|---|
| **Base URL** | Contact-based access |
| **Auth** | API key |
| **Data Categories** | (1) Visa & entry, (2) Safety & security, (3) Health risks, (4) Customs requirements, (5) General info |
| **Coverage** | 200+ countries |
| **Docs** | https://visadb.io/api |

---

## 4. Cost of Living & Currency

### 4.1 ExchangeRate-API

| Field | Value |
|---|---|
| **Base URL (open)** | `https://open.er-api.com/v6/latest/{base}` |
| **Base URL (keyed)** | `https://v6.exchangerate-api.com/v6/{key}` |
| **Auth** | None (open access) or API key in URL path |
| **Rate Limits** | Open: ~1 req/hour safe; Free plan: 1,500 req/month |
| **Currencies** | 161 (ISO 4217 codes) |
| **Response Format** | JSON |
| **Get API Key** | https://www.exchangerate-api.com (free plan available) |

**Key Endpoints:**

```
# Open access (no key)
GET https://open.er-api.com/v6/latest/USD

# Standard (with key)
GET https://v6.exchangerate-api.com/v6/{key}/latest/USD

# Pair conversion
GET https://v6.exchangerate-api.com/v6/{key}/pair/USD/EUR

# Pair with amount
GET https://v6.exchangerate-api.com/v6/{key}/pair/USD/EUR/100

# Supported codes
GET https://v6.exchangerate-api.com/v6/{key}/codes
```

**Response Structure (standard):**
```json
{
  "result": "success",
  "base_code": "USD",
  "time_last_update_utc": "Mon, 16 Feb 2026 00:00:01 +0000",
  "time_next_update_utc": "Tue, 17 Feb 2026 00:00:01 +0000",
  "conversion_rates": {
    "EUR": 0.9234,
    "GBP": 0.7891,
    "JPY": 149.52
  }
}
```

---

### 4.2 Numbeo API (Paid)

| Field | Value |
|---|---|
| **Base URL** | `https://www.numbeo.com/api/` |
| **Auth** | `api_key` query parameter |
| **Rate Limits** | Included with subscription |
| **Pricing** | ~$220–260/month (not free) |
| **Response Format** | JSON |
| **Docs** | https://www.numbeo.com/common/api.jsp |

**Key Endpoints:**

```
# City prices
GET /api/city_prices?api_key={key}&query=Paris,France

# Country prices
GET /api/country_prices?api_key={key}&country=France

# Cost of living indices
GET /api/indices?api_key={key}&query=Paris,France

# City rankings
GET /api/rankings_by_city_current?api_key={key}&section=8
```

Returns: restaurant prices, grocery prices, rent, transportation costs, utilities, cost-of-living index, purchasing power index, safety index, healthcare index.

**Note:** Numbeo is paid. For a free alternative, consider scraping the public Numbeo website or using general LLM knowledge for cost comparisons.

---

## 5. Flights & Travel Costs

### 5.1 Amadeus — Flight Offers Search v2

| Field | Value |
|---|---|
| **Base URL** | `https://test.api.amadeus.com` (test) / `https://api.amadeus.com` (prod) |
| **Auth** | OAuth2 Bearer token (see [Amadeus Auth](#amadeus-authentication)) |
| **Rate Limits** | 10 TPS (test env) |
| **Response Format** | JSON |

```
# One-way search
GET /v2/shopping/flight-offers?originLocationCode=PAR&destinationLocationCode=LON&departureDate=2026-06-15&adults=2&max=5&currencyCode=EUR

# Round-trip
GET /v2/shopping/flight-offers?originLocationCode=NYC&destinationLocationCode=PAR&departureDate=2026-06-15&returnDate=2026-06-22&adults=1&travelClass=BUSINESS&nonStop=true&max=3
```

**Key Parameters:** `originLocationCode`, `destinationLocationCode`, `departureDate`, `returnDate`, `adults`, `children`, `infants`, `travelClass` (ECONOMY/PREMIUM_ECONOMY/BUSINESS/FIRST), `nonStop`, `currencyCode`, `maxPrice`, `max`.

---

### 5.2 Travelpayouts Data API

| Field | Value |
|---|---|
| **Base URL** | `https://api.travelpayouts.com` |
| **Auth** | `X-Access-Token: {token}` header |
| **Rate Limits** | 200 req/hour per IP, 60 req/min per method |
| **Response Format** | JSON |
| **Get Token** | https://www.travelpayouts.com (affiliate signup) |

**Key Endpoints:**

```
# Cheapest tickets (cached prices)
GET /aviasales/v3/prices_for_dates?origin=MOW&destination=PAR&departure_at=2026-06&return_at=2026-07&sorting=price&limit=5

# Popular destinations from origin
GET /aviasales/v3/get_special_offers?origin=LON&currency=EUR

# Monthly price calendar
GET /aviasales/v3/prices_for_dates?origin=NYC&destination=PAR&departure_at=2026-06&one_way=true
```

Good for "cheapest time to fly" insights. Cached prices — not real-time booking.

---

## 6. Events & Festivals

### 6.1 Ticketmaster Discovery API v2

| Field | Value |
|---|---|
| **Base URL** | `https://app.ticketmaster.com/discovery/v2` |
| **Auth** | API key via `apikey` query parameter |
| **Rate Limits** | 5,000 calls/day, 2 req/sec (free tier) |
| **Coverage** | 230,000+ events from Ticketmaster, Universe, FrontGate |
| **Response Format** | HAL+JSON |
| **Get API Key** | https://developer.ticketmaster.com/ (free registration) |

**Key Endpoints:**

```
# Search events
GET /v2/events.json?apikey={key}&city=Paris&countryCode=FR&startDateTime=2026-06-01T00:00:00Z&endDateTime=2026-06-30T23:59:59Z&classificationName=music&size=20

# Search by geo
GET /v2/events.json?apikey={key}&geoPoint=48.8566,2.3522&radius=25&unit=km&size=20

# Event details
GET /v2/events/{eventId}.json?apikey={key}

# Venue search
GET /v2/venues.json?apikey={key}&city=London&countryCode=GB&size=10

# Attraction search
GET /v2/attractions.json?apikey={key}&keyword=Coldplay&size=5

# Classifications (genres/segments)
GET /v2/classifications.json?apikey={key}
```

**Key Query Parameters:**

| Parameter | Description |
|---|---|
| `keyword` | Free-text search |
| `city` | City name |
| `countryCode` | ISO country code |
| `stateCode` | US state code |
| `startDateTime` | ISO 8601 format |
| `endDateTime` | ISO 8601 format |
| `classificationName` | `music`, `sports`, `arts`, `theatre`, `family`, `film` |
| `geoPoint` | `lat,lon` |
| `radius` | Search radius |
| `unit` | `miles` or `km` |
| `size` | Results per page (max 200) |
| `page` | Page number |
| `sort` | `date,asc`, `date,desc`, `name,asc`, `relevance,desc` |

**Response Structure:**
```json
{
  "_embedded": {
    "events": [
      {
        "name": "Concert Name",
        "id": "Z7r9jZ1A...",
        "dates": {
          "start": {"localDate": "2026-06-15", "localTime": "20:00:00"},
          "status": {"code": "onsale"}
        },
        "classifications": [{"segment": {"name": "Music"}, "genre": {"name": "Rock"}}],
        "priceRanges": [{"min": 45.0, "max": 120.0, "currency": "EUR"}],
        "_embedded": {
          "venues": [{"name": "AccorHotels Arena", "city": {"name": "Paris"}}]
        },
        "url": "https://www.ticketmaster.com/event/..."
      }
    ]
  },
  "page": {"size": 20, "totalElements": 150, "totalPages": 8, "number": 0}
}
```

---

## 7. Country & Destination Metadata

### 7.1 REST Countries API

| Field | Value |
|---|---|
| **Base URL** | `https://restcountries.com/v3.1` |
| **Auth** | None required |
| **Rate Limits** | No published limits (open source, ~4M hits/day) |
| **Response Format** | JSON array |
| **Docs** | https://restcountries.com |

**Key Endpoints:**

```
# All countries (fields REQUIRED)
GET /v3.1/all?fields=name,capital,currencies,region,population,flags,languages,timezones,latlng

# By name
GET /v3.1/name/france
GET /v3.1/name/france?fullText=true

# By code
GET /v3.1/alpha/FR
GET /v3.1/alpha?codes=FR,DE,IT,ES

# By region
GET /v3.1/region/europe

# By currency
GET /v3.1/currency/eur

# By language
GET /v3.1/lang/french

# By capital
GET /v3.1/capital/paris
```

**The `?fields=` parameter** is comma-separated and **mandatory for `/all`**. Available fields (30+): `name`, `capital`, `currencies`, `languages`, `region`, `subregion`, `population`, `latlng`, `borders`, `timezones`, `flags`, `maps`, `area`, `demonyms`, `car`, `gini`, `cca2`, `cca3`, `tld`, `idd`, `capitalInfo`, `startOfWeek`, `postalCode`, `coatOfArms`, `continents`.

**Response Structure (key fields):**
```json
{
  "name": {"common": "France", "official": "French Republic"},
  "capital": ["Paris"],
  "region": "Europe",
  "subregion": "Western Europe",
  "population": 67390000,
  "languages": {"fra": "French"},
  "currencies": {"EUR": {"name": "Euro", "symbol": "€"}},
  "timezones": ["UTC-10:00", "UTC+01:00", "UTC+12:00"],
  "latlng": [46.0, 2.0],
  "borders": ["AND", "BEL", "DEU", "ITA", "LUX", "MCO", "ESP", "CHE"],
  "flags": {"png": "https://flagcdn.com/w320/fr.png", "svg": "https://flagcdn.com/fr.svg"},
  "capitalInfo": {"latlng": [48.87, 2.33]},
  "car": {"side": "right"},
  "startOfWeek": "monday"
}
```

---

### 7.2 Amadeus — Travel Recommendations

```
GET /v1/reference-data/recommended-locations?cityCodes=PAR&travelerCountryCode=US
```

AI-powered destination suggestions based on traveler profile and origin city.

---

## 8. Air Quality & Environmental

### 8.1 Open-Meteo Air Quality API

| Field | Value |
|---|---|
| **Base URL** | `https://air-quality-api.open-meteo.com/v1/air-quality` |
| **Auth** | None required |
| **Rate Limits** | Same as Open-Meteo weather (10K/day, 5K/hour, 600/min) |
| **Response Format** | JSON |
| **License** | CC-BY 4.0 |

```
GET /v1/air-quality?latitude=48.85&longitude=2.35&hourly=pm10,pm2_5,us_aqi,european_aqi&forecast_days=3
```

**Available Variables:** `pm10`, `pm2_5`, `carbon_monoxide`, `nitrogen_dioxide`, `sulphur_dioxide`, `ozone`, `us_aqi`, `european_aqi`, `alder_pollen`, `birch_pollen`, `grass_pollen`, `mugwort_pollen`, `olive_pollen`, `ragweed_pollen`.

---

### 8.2 WAQI (World Air Quality Index)

| Field | Value |
|---|---|
| **Base URL** | `https://api.waqi.info` |
| **Auth** | Token via `token` query parameter |
| **Rate Limits** | 1,000 req/min (burst: 60/sec) |
| **Coverage** | 12,000+ monitoring stations worldwide |
| **Response Format** | JSON |
| **Get Token** | https://aqicn.org/data-platform/token/ (free) |

**Key Endpoints:**

```
# By city name
GET /feed/paris/?token={token}

# By geo coordinates
GET /feed/geo:48.8566;2.3522/?token={token}

# Search stations
GET /search/?keyword=paris&token={token}

# Map bounds (for visualization)
GET /map/bounds/?token={token}&latlng=48.8,2.2,48.9,2.4
```

**Response Structure:**
```json
{
  "status": "ok",
  "data": {
    "aqi": 42,
    "idx": 5722,
    "city": {"name": "Paris", "geo": [48.8566, 2.3522]},
    "dominentpol": "pm25",
    "iaqi": {
      "pm25": {"v": 42},
      "pm10": {"v": 28},
      "o3": {"v": 15},
      "no2": {"v": 22},
      "t": {"v": 15.3},
      "h": {"v": 72}
    },
    "time": {"s": "2026-02-16 14:00:00"}
  }
}
```

AQI scale: 0–50 (Good), 51–100 (Moderate), 101–150 (Unhealthy for Sensitive), 151–200 (Unhealthy), 201–300 (Very Unhealthy), 300+ (Hazardous).

---

## 9. Destination Photos

### 9.1 Unsplash API

| Field | Value |
|---|---|
| **Base URL** | `https://api.unsplash.com` |
| **Auth** | `Authorization: Client-ID {access_key}` header |
| **Rate Limits** | Demo: 50 req/hour; Production: 5,000 req/hour |
| **Response Format** | JSON |
| **License** | Free for all uses with attribution |
| **Get API Key** | https://unsplash.com/developers (register app) |

**Key Endpoints:**

```
# Search photos by query
GET /search/photos?query=Paris+travel&per_page=10&orientation=landscape

# Get random photo
GET /photos/random?query=Tokyo+skyline&count=1

# Get photo by ID
GET /photos/{id}
```

**Response Structure (search):**
```json
{
  "total": 5000,
  "total_pages": 500,
  "results": [
    {
      "id": "abc123",
      "description": "Eiffel Tower at sunset",
      "urls": {
        "raw": "https://images.unsplash.com/photo-...",
        "full": "https://images.unsplash.com/photo-...?w=4000",
        "regular": "https://images.unsplash.com/photo-...?w=1080",
        "small": "https://images.unsplash.com/photo-...?w=400",
        "thumb": "https://images.unsplash.com/photo-...?w=200"
      },
      "links": {"html": "https://unsplash.com/photos/abc123"},
      "user": {"name": "Photographer Name", "links": {"html": "..."}}
    }
  ]
}
```

**Attribution required:** Link to photo and photographer per Unsplash guidelines.

---

### 9.2 Pexels API

| Field | Value |
|---|---|
| **Base URL** | `https://api.pexels.com/v1` |
| **Auth** | `Authorization: {API_KEY}` header |
| **Rate Limits** | 200 req/hour, 20,000 req/month (default free) |
| **Response Format** | JSON |
| **License** | Free, no attribution required |
| **Get API Key** | https://www.pexels.com/api/ (free registration) |

**Key Endpoints:**

```
# Search photos
GET /v1/search?query=Bali+beach+travel&per_page=10&orientation=landscape

# Curated photos
GET /v1/curated?per_page=15&page=1

# Get photo by ID
GET /v1/photos/{id}
```

**Response Structure:**
```json
{
  "total_results": 10000,
  "page": 1,
  "per_page": 10,
  "photos": [
    {
      "id": 12345,
      "width": 4000,
      "height": 2667,
      "url": "https://www.pexels.com/photo/...",
      "photographer": "Name",
      "photographer_url": "https://www.pexels.com/@name",
      "src": {
        "original": "https://images.pexels.com/photos/12345/...",
        "large2x": "https://images.pexels.com/photos/12345/...?w=1880",
        "large": "https://images.pexels.com/photos/12345/...?w=940",
        "medium": "https://images.pexels.com/photos/12345/...?w=350",
        "small": "https://images.pexels.com/photos/12345/...?w=130"
      },
      "alt": "Beach in Bali at sunset"
    }
  ]
}
```

---

## Quick Reference: API Keys Needed

| API | Key Required | How to Get |
|-----|-------------|------------|
| Open-Meteo (weather + AQ) | No | — |
| REST Countries | No | — |
| travel-advisory.info | No | — |
| US State Dept | No | — |
| ExchangeRate-API (open) | No | — |
| OpenTripMap | Yes (free) | https://opentripmap.io/product |
| OpenWeatherMap | Yes (free) | https://home.openweathermap.org/users/sign_up |
| WeatherAPI.com | Yes (free) | https://www.weatherapi.com/signup.aspx |
| Ticketmaster | Yes (free) | https://developer.ticketmaster.com/ |
| TuGo | Yes (free) | https://developer.tugo.com |
| WAQI | Yes (free) | https://aqicn.org/data-platform/token/ |
| Unsplash | Yes (free) | https://unsplash.com/developers |
| Pexels | Yes (free) | https://www.pexels.com/api/ |
| Amadeus | Yes (free tier) | https://developers.amadeus.com/register |
| Geoapify | Yes (free tier) | https://myprojects.geoapify.com/ |
| Foursquare | Yes (free tier) | https://foursquare.com/developers/signup |
| ExchangeRate-API (keyed) | Yes (free plan) | https://www.exchangerate-api.com |
| Travelpayouts | Yes (affiliate) | https://www.travelpayouts.com |
| Numbeo | Yes (paid ~$250/mo) | https://www.numbeo.com/common/api.jsp |
