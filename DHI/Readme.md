# Weather API Service

REST API for weather data with persistent caching and graceful degradation.

## Overview

Provides current weather and hourly forecasts for cities. Uses Open-Meteo as data source with disk-based cache to reduce bandwidth and maintain availability during network failures.

## Features

- Persistent cache (survives restarts)
- Stale data fallback (up to 24hrs when API unavailable)
- Configurable TTL (current: 30min, hourly: 1hr)
- Graceful shutdown with cache preservation
- Thread-safe concurrent requests
- Automatic geocoding

## Installation
```bash
git clone <repository-url>
cd weather-api
go run .
```

Runs on `http://localhost:8080`

---

**Parameters:**
- `city`: City name
- `start_date`, `end_date`: Date range (YYYY-MM-DD)
- `data_type`: `current`, `hourly`, or `both`

```

## Configuration

**Ports:** `DHI.config.go`  
**Cache TTL:** `weather.go`  
**Stale window:** `cache.go` (default: 24hrs)

## Dependencies

- Go 1.21+
- Open-Meteo API
- Open-Meteo Geocoding API

---
## Example Usage
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "SrID": "weather.forecast",
    "Seed": {
      "city": "Lagos",
      "start_date": "2025-12-17",
      "end_date": "2025-12-24",
      "data_type": "current"
    }
  }'