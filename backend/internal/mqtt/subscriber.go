// Package mqtt provides MQTT message subscription for IoT sensor data.
// It parses incoming JSON payloads from sensors and forwards validated
// EnvironmentReading values to a handler callback.
package mqtt

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
)

// SensorPayload represents the raw JSON payload received from an IoT sensor via MQTT.
type SensorPayload struct {
	SensorID    string   `json:"sensor_id"`
	Location    string   `json:"location"`
	Temperature *float64 `json:"temperature"`
	Humidity    *float64 `json:"humidity"`
	Timestamp   string   `json:"timestamp,omitempty"`
}

// Validate checks that the payload contains all required fields and valid values.
func (p *SensorPayload) Validate() error {
	if p.SensorID == "" {
		return errors.New("sensor_id is required")
	}
	if p.Location == "" {
		return errors.New("location is required")
	}
	if p.Temperature == nil {
		return errors.New("temperature is required")
	}
	if p.Humidity == nil {
		return errors.New("humidity is required")
	}
	if *p.Temperature < domain.TemperatureMin || *p.Temperature > domain.TemperatureMax {
		return fmt.Errorf("temperature must be between %.1f and %.1f", domain.TemperatureMin, domain.TemperatureMax)
	}
	if *p.Humidity < domain.HumidityMin || *p.Humidity > domain.HumidityMax {
		return fmt.Errorf("humidity must be between %.1f and %.1f", domain.HumidityMin, domain.HumidityMax)
	}
	return nil
}

// ToReading converts a validated SensorPayload to a domain.EnvironmentReading.
// If Timestamp is empty or unparseable, the current UTC time is used.
func (p *SensorPayload) ToReading() domain.EnvironmentReading {
	t := time.Now().UTC()
	if p.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, p.Timestamp); err == nil {
			t = parsed.UTC()
		}
	}
	return domain.EnvironmentReading{
		Time:        t,
		SensorID:    p.SensorID,
		Location:    p.Location,
		Temperature: *p.Temperature,
		Humidity:    *p.Humidity,
	}
}

// MessageHandler is a callback that processes a validated EnvironmentReading.
type MessageHandler func(reading domain.EnvironmentReading) error

// Subscriber listens to an MQTT topic and processes sensor messages.
// In this MVP, it provides ParseMessage for integration with any MQTT client library.
type Subscriber struct {
	topic   string
	handler MessageHandler
}

// NewSubscriber creates a new MQTT Subscriber for the given topic.
func NewSubscriber(topic string, handler MessageHandler) *Subscriber {
	return &Subscriber{
		topic:   topic,
		handler: handler,
	}
}

// Topic returns the MQTT topic this subscriber listens to.
func (s *Subscriber) Topic() string {
	return s.topic
}

// ParseMessage parses and validates a raw MQTT message payload, then forwards
// the resulting EnvironmentReading to the handler.
// Returns an error if the payload is invalid or the handler fails.
func (s *Subscriber) ParseMessage(payload []byte) error {
	var sensorPayload SensorPayload
	if err := json.Unmarshal(payload, &sensorPayload); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	if err := sensorPayload.Validate(); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}

	reading := sensorPayload.ToReading()
	return s.handler(reading)
}
