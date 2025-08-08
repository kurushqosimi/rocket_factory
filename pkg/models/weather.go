package models

import "time"

// Weather presents info about a weather of specific city
type Weather struct {
	// Name of the city
	City string `json:"city"`
	// Temperature in Celsius
	Temperature float64 `json:"temperature"`
	// Last updated time
	UpdatedAt time.Time `json:"updated_at"`
}
