package models

import "sync"

// WeatherStorage is a thread safe storage for weather
type WeatherStorage struct {
	mu       sync.RWMutex
	weathers map[string]*Weather
}

// NewWeatherStorage creates a new storage for weather data
func NewWeatherStorage() *WeatherStorage {
	return &WeatherStorage{
		weathers: make(map[string]*Weather),
	}
}

// GetWeather returns a data about a weather by the city name
// if not found returns nil
func (s *WeatherStorage) GetWeather(city string) *Weather {
	s.mu.RLock()
	defer s.mu.RUnlock()

	weather, ok := s.weathers[city]
	if !ok {
		return nil
	}

	return weather
}

// UpdateWeather updates weather information for the given city
// if it does not exist creates one
func (s *WeatherStorage) UpdateWeather(weather *Weather) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.weathers[weather.City] = weather
}
