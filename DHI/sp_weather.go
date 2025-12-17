package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type WeatherAPIResponse struct {
	Current struct {
		Temperature      float64 `json:"temperature_2m"`
		RelativeHumidity float64 `json:"relative_humidity_2m"`
		ApparentTemp     float64 `json:"apparent_temperature"`
		IsDay            int     `json:"is_day"`
		Precipitation    float64 `json:"precipitation"`
		CloudCover       float64 `json:"cloud_cover"`
		WindSpeed        float64 `json:"wind_speed_10m"`
		WindDirection    float64 `json:"wind_direction_10m"`
		WindGusts        float64 `json:"wind_gusts_10m"`
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
	Daily struct {
		Time             []string  `json:"time"`
		CloudCoverMean   []float64 `json:"cloud_cover_mean"`
		TempMean         []float64 `json:"temperature_2m_mean"`
		WindDirectionDom []float64 `json:"winddirection_10m_dominant"`
		WindSpeedMean    []float64 `json:"wind_speed_10m_mean"`
	} `json:"daily"`
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

	// 1. Extract inputs from Seed
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

	// 2. Geocode the city to get coordinates
	latitude, longitude, err := geocodeCity(city)
	if err != nil {
		return 400, fmt.Sprintf("failed to geocode city: %s", err.Error()), nil
	}

	// 3. Build Open-Meteo URL
	weatherURL := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast"+
			"?latitude=%f"+
			"&longitude=%f"+
			"&current=temperature_2m,relative_humidity_2m,apparent_temperature,is_day,wind_speed_10m,wind_direction_10m,precipitation,cloud_cover,wind_gusts_10m"+
			"&hourly=temperature_2m,relative_humidity_2m,precipitation,cloud_cover,wind_speed_10m,sunshine_duration"+
			"&daily=cloud_cover_mean,temperature_2m_mean,winddirection_10m_dominant,wind_speed_10m_mean"+
			"&timezone=auto"+
			"&start_date=%s"+
			"&end_date=%s",
		latitude,
		longitude,
		startDate,
		endDate,
	)

	// 4. Call external API
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(weatherURL)
	if err != nil {
		return 502, "failed to contact weather service", nil
	}
	defer resp.Body.Close()

	// 5. Parse the response
	var apiResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 500, "failed to parse weather response", nil
	}

	// 6. Return structured data
	return 200, "Weather data retrieved successfully", map[string]any{
		"city":     city,
		"location": map[string]float64{"latitude": latitude, "longitude": longitude},
		"current":  apiResp.Current,
		"hourly":   apiResp.Hourly,
		"daily":    apiResp.Daily,
	}
}