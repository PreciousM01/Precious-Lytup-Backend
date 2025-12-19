# DaemonCore + DHI Weather API

A modular Go backend combining **DaemonCore** lifecycle management with **DHI (Daemon HTTP Interface)** to run a weather forecast API service with persistent caching and graceful degradation.

## Architecture

### DaemonCore
Manages daemon lifecycle:
- Starts and supervises long-running processes
- Tracks state (startup, running, shutdown)
- Coordinates graceful shutdown via OS signals
- Captures execution results and errors
- Ensures controlled concurrency

### DHI (Daemon HTTP Interface)
HTTP server implemented as a daemon:
- HTTP/HTTPS servers with routing
- Centralized request handling
- Panic recovery and response normalization
- Service Provider (SP) abstraction for extensible APIs
- Runs as a managed daemon under DaemonCore

### Weather Service Provider
Business logic running in DHI:
- Geocoding city names to coordinates
- Weather data retrieval from Open-Meteo API
- Persistent disk-based cache
- Stale data fallback (up to 24hrs)
- Configurable TTL (current: 30min, hourly: 1hr)

## Key Features

**Daemon Management:**
- Concurrent daemon execution with goroutines
- Thread-safe state management with mutexes
- Graceful shutdown on OS signals (SIGINT, SIGTERM, SIGHUP)
- Configurable startup/shutdown grace periods

**HTTP Interface:**
- Service Provider routing pattern
- Request/response normalization
- Automatic panic recovery
- JSON API responses

**Weather Service:**
- Persistent cache (survives restarts)
- Graceful degradation (serves stale data when API unavailable)
- Automatic geocoding
- Thread-safe cache operations

## Structure
```
DHI/
├── Main.go              # DaemonCore entry point
├── DHI-go-G1.go         # HTTP interface daemon
├── sp_weather.go        # Weather Service Provider
├── cache.go             # Persistent cache manager
├── Test.go              # Service registration
├── Main.conf.go         # Daemon configuration
├── DHI-go-G1.conf.go    # Server configuration (ports, TLS, etc.)
├── weather_cache.json   # Cache storage (auto-generated)
├── tls.crt, tls.key     # TLS certificates
├── go.mod               # Go dependencies
└── LICENSE
```

## Installation
```bash
git clone <repository-url>
cd weather-api
go run .
```
Server runs on `http://localhost:8080`

**Parameters:**
- `city`: City name
- `start_date`, `end_date`: Date range (YYYY-MM-DD)
- `data_type`: `current`, `hourly`, or `both`

## Request Flow
```
Client Request
    ↓
DHI HTTP Server
    ↓
Service Router
    ↓
Weather Service Provider
    ↓
Cache Check → API Call
    ↓
Response
```

## Lifecycle

**Startup:**
1. DaemonCore initializes
2. Load cache from disk
3. DHI daemon starts HTTP servers
4. Weather SP registers with DHI
5. Ready to serve requests

**Shutdown (Ctrl+C):**
1. OS signal received
2. Cache saved to disk
3. DHI servers closed
4. Daemon cleanup
5. Process exit

## Configuration

**Server Ports:** `DHI.config.go`  
**Cache TTL:** `weather.go`  
**Service Registration:** `test.go`  
**Daemon Settings:** `main.config.go`

## Tech Stack

- **Language:** Go
- **Concurrency:** Goroutines, channels, mutexes
- **Networking:** `net/http`
- **Process Control:** `os/signal`, `syscall`
- **Architecture:** Daemon-based lifecycle + Service Provider pattern
- **External APIs:** Open-Meteo (weather data), Open-Meteo Geocoding

## Dependencies

- Go 1.21+
- Open-Meteo API
- Open-Meteo Geocoding API

## Extending

Add new Service Providers in `test.go`:
```go
DHI0_SPRegister = []*DHI0_SP{
    {Code: "weather.forecast", Program: SPWeatherForecast},
    {Code: "your.service", Program: YourServiceFunction},
}
```

Service Provider signature:
```go
func YourService(r *http.Request, srID string, seed map[string]any) (code int, note string, yield any)
```
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
```

## License

MIT License