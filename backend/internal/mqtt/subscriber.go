// Package mqtt provides MQTT message subscription for IoT sensor data.
// It parses incoming JSON payloads from sensors and forwards validated
// EnvironmentReading values to a handler callback.
package mqtt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
)

const (
	// DefaultMaxPayloadSize is the default maximum accepted MQTT payload size
	// (256 KiB). Sensor JSON payloads are well under 1 KiB, so this leaves ample
	// headroom while preventing memory exhaustion from oversized messages.
	DefaultMaxPayloadSize = 256 * 1024

	// MaxPayloadSizeEnvVar overrides DefaultMaxPayloadSize when set to a positive integer.
	MaxPayloadSizeEnvVar = "MQTT_MAX_PAYLOAD_BYTES"
)

// ErrPayloadTooLarge is returned by ParseMessage when a payload exceeds the
// configured maximum size. Callers should discard the message and log the event.
var ErrPayloadTooLarge = errors.New("mqtt payload exceeds maximum allowed size")

// MaxPayloadSizeFromEnv returns the maximum payload size configured via the
// MQTT_MAX_PAYLOAD_BYTES environment variable, or DefaultMaxPayloadSize when the
// variable is unset, non-numeric, or non-positive.
func MaxPayloadSizeFromEnv() int {
	raw := os.Getenv(MaxPayloadSizeEnvVar)
	if raw == "" {
		return DefaultMaxPayloadSize
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return DefaultMaxPayloadSize
	}
	return n
}

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
	topic          string
	handler        MessageHandler
	maxPayloadSize int
}

// NewSubscriber creates a new MQTT Subscriber for the given topic using the
// default maximum payload size.
func NewSubscriber(topic string, handler MessageHandler) *Subscriber {
	return NewSubscriberWithMaxPayload(topic, handler, DefaultMaxPayloadSize)
}

// NewSubscriberWithMaxPayload creates a Subscriber with an explicit maximum
// payload size in bytes. A non-positive value falls back to DefaultMaxPayloadSize.
func NewSubscriberWithMaxPayload(topic string, handler MessageHandler, maxPayloadSize int) *Subscriber {
	if maxPayloadSize <= 0 {
		maxPayloadSize = DefaultMaxPayloadSize
	}
	return &Subscriber{
		topic:          topic,
		handler:        handler,
		maxPayloadSize: maxPayloadSize,
	}
}

// MaxPayloadSize returns the maximum payload size, in bytes, this subscriber accepts.
func (s *Subscriber) MaxPayloadSize() int {
	return s.maxPayloadSize
}

// Topic returns the MQTT topic this subscriber listens to.
func (s *Subscriber) Topic() string {
	return s.topic
}

// ParseMessage parses and validates a raw MQTT message payload, then forwards
// the resulting EnvironmentReading to the handler.
// Returns ErrPayloadTooLarge if the payload exceeds the configured maximum size,
// or an error if the payload is invalid or the handler fails.
func (s *Subscriber) ParseMessage(payload []byte) error {
	if len(payload) > s.maxPayloadSize {
		return fmt.Errorf("%w: %d bytes (limit %d)", ErrPayloadTooLarge, len(payload), s.maxPayloadSize)
	}

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
