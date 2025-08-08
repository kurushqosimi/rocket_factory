package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/kurushqosimi/rocket_factory/pkg/models"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	httpPort     = "8080"
	urlParamCity = "city"

	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	// storage init
	storage := models.NewWeatherStorage()

	// router initialization
	r := chi.NewRouter()

	// middleware adding
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// routes
	r.Route("/api/v1/weather", func(r chi.Router) {
		r.Get("/{city}", getWeatherHandler(storage))
		r.Put("/{city}", updateWeatherHandler(storage))
	})

	// HTTP server start
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// server start on separate goroutine
	go func() {
		log.Printf("HTTP server started on port %s\n", httpPort)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed to start a server: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting server down...")

	// context with timeout for server stop
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("Error trying server shut down: %v\n", err)
	}

	log.Println("Server stopped")
}

// getWeatherHandler processes requests for getting information about weather for the specific city
func getWeatherHandler(storage *models.WeatherStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, urlParamCity)
		if city == "" {
			http.Error(w, "City parameter is required", http.StatusBadRequest)
			return
		}

		weather := storage.GetWeather(city)
		if weather == nil {
			http.Error(w, fmt.Sprintf("Weather for city '%s' not found", city), http.StatusNotFound)
			return
		}

		render.JSON(w, r, weather)
	}
}

// updateWeatherHandler processes requests for updating information about weather for the specific city
func updateWeatherHandler(storage *models.WeatherStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, urlParamCity)
		if city == "" {
			http.Error(w, "City parameter is required", http.StatusBadRequest)
			return
		}

		// decoding data from request body
		var weatherUpdate models.Weather
		if err := json.NewDecoder(r.Body).Decode(&weatherUpdate); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// installs city name from url param
		weatherUpdate.City = city

		// installs update time
		weatherUpdate.UpdatedAt = time.Now()

		// updates info about the weather
		storage.UpdateWeather(&weatherUpdate)

		// returns updated data
		render.JSON(w, r, weatherUpdate)
	}
}
