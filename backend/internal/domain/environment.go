package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	// TemperatureMin is the minimum valid temperature in Celsius.
	TemperatureMin = -10.0
	// TemperatureMax is the maximum valid temperature in Celsius.
	TemperatureMax = 100.0
	// HumidityMin is the minimum valid humidity percentage.
	HumidityMin = 0.0
	// HumidityMax is the maximum valid humidity percentage.
	HumidityMax = 100.0
)

// EnvironmentReading represents a single sensor measurement from the urushi-buro.
type EnvironmentReading struct {
	Time          time.Time  `json:"time"`
	SensorID      string     `json:"sensor_id"`
	Location      string     `json:"location"`
	Temperature   float64    `json:"temperature"`
	Humidity      float64    `json:"humidity"`
	WorkID        *uuid.UUID `json:"work_id,omitempty"`
	ProcessStepID *uuid.UUID `json:"process_step_id,omitempty"`
}

// Validate checks the environment reading's fields for correctness.
func (er *EnvironmentReading) Validate() error {
	if er.SensorID == "" {
		return errors.New("sensor_id must not be empty")
	}
	if er.Location == "" {
		return errors.New("location must not be empty")
	}
	if er.Temperature < TemperatureMin || er.Temperature > TemperatureMax {
		return errors.New("temperature must be between -10.0 and 100.0")
	}
	if er.Humidity < HumidityMin || er.Humidity > HumidityMax {
		return errors.New("humidity must be between 0.0 and 100.0")
	}
	if er.Time.IsZero() {
		return errors.New("time must not be zero")
	}
	return nil
}

// AlertThreshold defines the acceptable ranges for environmental conditions.
type AlertThreshold struct {
	ID             uuid.UUID `json:"id"`
	SensorID       string    `json:"sensor_id"`
	TemperatureMin float64   `json:"temperature_min"`
	TemperatureMax float64   `json:"temperature_max"`
	HumidityMin    float64   `json:"humidity_min"`
	HumidityMax    float64   `json:"humidity_max"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Validate checks the alert threshold's fields for correctness.
func (at *AlertThreshold) Validate() error {
	if at.SensorID == "" {
		return errors.New("sensor_id must not be empty")
	}
	if at.TemperatureMin >= at.TemperatureMax {
		return errors.New("temperature_min must be less than temperature_max")
	}
	if at.HumidityMin >= at.HumidityMax {
		return errors.New("humidity_min must be less than humidity_max")
	}
	if at.TemperatureMin < TemperatureMin || at.TemperatureMax > TemperatureMax {
		return errors.New("temperature range must be within -10.0 to 100.0")
	}
	if at.HumidityMin < HumidityMin || at.HumidityMax > HumidityMax {
		return errors.New("humidity range must be within 0.0 to 100.0")
	}
	return nil
}

// IsOutOfRange checks whether the given reading exceeds this threshold.
func (at *AlertThreshold) IsOutOfRange(reading EnvironmentReading) bool {
	if !at.Enabled {
		return false
	}
	return reading.Temperature < at.TemperatureMin ||
		reading.Temperature > at.TemperatureMax ||
		reading.Humidity < at.HumidityMin ||
		reading.Humidity > at.HumidityMax
}
