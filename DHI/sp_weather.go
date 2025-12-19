package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type WeatherAPIResponse struct {
	Current struct {
		Temperature      float64 `json:"temperature_2m"`
		RelativeHumidity float64 `json:"relative_humidity_2m"`
		ApparentTemp     float64 `json:"apparent_temperature"`
		Precipitation    float64 `json:"precipitation"`
		CloudCover       float64 `json:"cloud_cover"`
		WindSpeed        float64 `json:"wind_speed_10m"`
	} `json:"current"`
	Hourly struct {
		Time             []string  `json:"time"`
		Temperature      []float64 `json:"temperature_2m"`
		RelativeHumidity []float64 `json:"relative_humidity_2m"`
		Precipitation    []float64 `json:"precipitation"`
		CloudCover       []float64 `json:"cloud_cover"`
		WindSpeed        []float64 `json:"wind_speed_10m"`
		SunshineDuration []float64 `json:"sunshine_duration"`
	} `json:"hourly"`
}

type GeocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

// Helper function to geocode city name to coordinates
func geocodeCity(city string) (lat, lon float64, err error) {
	geocodeURL := fmt.Sprintf(
		"https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json",
		url.QueryEscape(city),
	)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(geocodeURL)
	if err != nil {
		return 0, 0, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer resp.Body.Close()

	var geoResp GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return 0, 0, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if len(geoResp.Results) == 0 {
		return 0, 0, fmt.Errorf("city not found: %s", city)
	}

	return geoResp.Results[0].Latitude, geoResp.Results[0].Longitude, nil
}

// Weather Service Provider
func SPWeatherForecast(
	r *http.Request,
	srID string,
	seed map[string]any,
) (C int, N string, Y any) {

	// 1. Extract inputs
	city, ok := seed["city"].(string)
	if !ok || city == "" {
		return 400, "missing city", nil
	}

	startDate, ok := seed["start_date"].(string)
	if !ok || startDate == "" {
		return 400, "missing start_date", nil
	}

	endDate, ok := seed["end_date"].(string)
	if !ok || endDate == "" {
		return 400, "missing end_date", nil
	}

	// Get data type (defaults to "both")
	dataType, _ := seed["data_type"].(string)
	if dataType == "" {
		dataType = "both"
	}

	// 2. Determine cache TTL based on data type
	var cacheTTL time.Duration
	switch dataType {
	case "current":
		cacheTTL = 30 * time.Minute
	case "hourly":
		cacheTTL = 1 * time.Hour
	default:
		cacheTTL = 30 * time.Minute
	}

	// 3. Check cache
	cacheKey := GlobalWeatherCache.GenerateKey(city+dataType, startDate, endDate)
	if cachedData, found := GlobalWeatherCache.Get(cacheKey); found {
		Output_Logg("OUT", "Weather", fmt.Sprintf("Cache HIT for %s (%s data)", city, dataType))
		return 200, "Weather data retrieved from cache", cachedData
	}

	Output_Logg("OUT", "Weather", fmt.Sprintf("Cache MISS for %s - fetching from API", city))

	// 4. Geocode
	latitude, longitude, err := geocodeCity(city)
	if err != nil {
		return 400, fmt.Sprintf("failed to geocode city: %s", err.Error()), nil
	}

	// 5. Build API URL
	weatherURL := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast"+
			"?latitude=%f"+
			"&longitude=%f"+
			"&current=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation,cloud_cover,wind_speed_10m"+
			"&hourly=temperature_2m,relative_humidity_2m,precipitation,cloud_cover,wind_speed_10m,sunshine_duration"+
			"&timezone=auto"+
			"&start_date=%s"+
			"&end_date=%s",
		latitude, longitude, startDate, endDate,
	)

	// 6. Fetch from API
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(weatherURL)
	if err != nil {
		return 502, "failed to contact weather service", nil
	}
	defer resp.Body.Close()

	var apiResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 500, "failed to parse weather response", nil
	}

	// 7. Build response based on data type
	responseData := map[string]any{
		"city":     city,
		"location": map[string]float64{"latitude": latitude, "longitude": longitude},
	}

	switch dataType {
	case "current":
		responseData["current"] = apiResp.Current
	case "hourly":
		responseData["hourly"] = apiResp.Hourly
	default: // "both"
		responseData["current"] = apiResp.Current
		responseData["hourly"] = apiResp.Hourly
	}

	// 8. Store in cache with dynamic TTL
	GlobalWeatherCache.SetWithTTL(cacheKey, responseData, cacheTTL)
	Output_Logg("OUT", "Weather", fmt.Sprintf("Cached %s data for %s (TTL: %v)", dataType, city, cacheTTL))

	return 200, "Weather data retrieved successfully", responseData
}