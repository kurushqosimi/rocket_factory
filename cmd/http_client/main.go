package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/kurushqosimi/rocket_factory/pkg/models"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	serverURL         = "http://localhost:8080"
	weatherAPIPath    = "/api/v1/weather/%s"
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	requestTimeout    = 5 * time.Second
	defaultCityName   = "Moscow"
	defaultMinTemp    = -10
	defaultMaxTemp    = 40
)

// httpClient - HTTP client with timeout
var httpClient = &http.Client{
	Timeout: requestTimeout,
}

// getWeather gets weather data for the specific city
func getWeather(ctx context.Context, city string) (*models.Weather, error) {
	// create request with context
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s"+weatherAPIPath, serverURL, city),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("get request create: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing get request: %w", err)
	}
	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			log.Printf("close resp body error: %v\n", err)
			return
		}
	}()

	// Not found check
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	// Ok check
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get weather data (status %d): %s", resp.StatusCode, defaultCityName)
	}

	// read request body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body error: %w", err)
	}

	var weather models.Weather
	err = json.Unmarshal(body, &weather)
	if err != nil {
		return nil, fmt.Errorf("JSON decode: %w", err)
	}

	return &weather, nil
}

// updateWeather update weather data for the specific city
func updateWeather(ctx context.Context, city string, weather *models.Weather) (*models.Weather, error) {
	// JSON Encode weather data
	jsonData, err := json.Marshal(weather)
	if err != nil {
		return nil, fmt.Errorf("JSON encode: %w", err)
	}

	// put request create
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s"+weatherAPIPath, serverURL, city),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("put request create: %w", err)
	}
	req.Header.Set(contentTypeHeader, contentTypeJSON)

	// req exec
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("put req exec: %w", err)
	}

	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			log.Printf("close resp body error: %v\n", err)
			return
		}
	}()

	// Ok check
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update weather data (status %d): %s", resp.StatusCode, defaultCityName)
	}

	// read request body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body error: %w", err)
	}

	var updatedWeather models.Weather
	err = json.Unmarshal(body, &updatedWeather)
	if err != nil {
		return nil, fmt.Errorf("JSON decode: %w", err)
	}

	return &updatedWeather, nil
}

// generateRandomWeather creates random weather data
func generateRandomWeather() *models.Weather {
	return &models.Weather{
		Temperature: gofakeit.Float64Range(defaultMinTemp, defaultMaxTemp),
	}
}

func main() {
	ctx := context.Background()

	log.Println("=== Testing Weather API ===")
	log.Println()

	log.Printf("Getting weather info for %s city", defaultCityName)
	log.Println("=======================================")

	// 1. Trying to get info about the city which does not exist yet
	weather, err := getWeather(ctx, defaultCityName)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Printf("Weather data for %s city: %+v\n", defaultCityName, weather)

	// 2. Updating weather data
	log.Printf("Updating weather data for %s city: %+v\n", defaultCityName, weather)
	log.Println("=======================================")

	newWeather := generateRandomWeather()

	updatedWeather, err := updateWeather(ctx, defaultCityName, newWeather)
	if err != nil {
		log.Printf("Update weather error: %v\n", err)
		return
	}
	log.Printf("Weather data updated: %+v\n", updatedWeather)

	// 3. Getting updated weather data
	log.Printf("Getting updated weather data for %s city\n", defaultCityName)
	log.Println("=======================================")

	weather, err = getWeather(ctx, defaultCityName)
	if err != nil {
		log.Printf("Get weather error: %v\n", err)
		return
	}

	if weather == nil {
		log.Printf("Unexpected: weather data does not exist after update: city %s\n", defaultCityName)
		return
	}

	log.Printf("Weather data: %+v\n", weather)
	log.Println("Testing finished successfully")
}
