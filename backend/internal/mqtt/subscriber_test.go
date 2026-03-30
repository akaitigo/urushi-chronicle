package mqtt_test

import (
	"strings"
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/mqtt"
)

func TestSensorPayload_Validate_Valid(t *testing.T) {
	temp := 25.0
	humidity := 75.0
	p := mqtt.SensorPayload{
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: &temp,
		Humidity:    &humidity,
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSensorPayload_Validate_MissingSensorID(t *testing.T) {
	temp := 25.0
	humidity := 75.0
	p := mqtt.SensorPayload{
		Location:    "urushi_buro",
		Temperature: &temp,
		Humidity:    &humidity,
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for missing sensor_id")
	}
}

func TestSensorPayload_Validate_MissingLocation(t *testing.T) {
	temp := 25.0
	humidity := 75.0
	p := mqtt.SensorPayload{
		SensorID:    "esp32-001",
		Temperature: &temp,
		Humidity:    &humidity,
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for missing location")
	}
}

func TestSensorPayload_Validate_MissingTemperature(t *testing.T) {
	humidity := 75.0
	p := mqtt.SensorPayload{
		SensorID: "esp32-001",
		Location: "urushi_buro",
		Humidity: &humidity,
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for missing temperature")
	}
}

func TestSensorPayload_Validate_MissingHumidity(t *testing.T) {
	temp := 25.0
	p := mqtt.SensorPayload{
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: &temp,
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for missing humidity")
	}
}

func TestSensorPayload_Validate_TemperatureOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		temp float64
	}{
		{"below minimum", -11.0},
		{"above maximum", 101.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			humidity := 75.0
			temp := tt.temp
			p := mqtt.SensorPayload{
				SensorID:    "esp32-001",
				Location:    "urushi_buro",
				Temperature: &temp,
				Humidity:    &humidity,
			}
			if err := p.Validate(); err == nil {
				t.Error("expected error for out-of-range temperature")
			}
		})
	}
}

func TestSensorPayload_Validate_HumidityOutOfRange(t *testing.T) {
	tests := []struct {
		name     string
		humidity float64
	}{
		{"below minimum", -1.0},
		{"above maximum", 101.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp := 25.0
			humidity := tt.humidity
			p := mqtt.SensorPayload{
				SensorID:    "esp32-001",
				Location:    "urushi_buro",
				Temperature: &temp,
				Humidity:    &humidity,
			}
			if err := p.Validate(); err == nil {
				t.Error("expected error for out-of-range humidity")
			}
		})
	}
}

func TestSensorPayload_ToReading(t *testing.T) {
	temp := 25.5
	humidity := 78.0
	p := mqtt.SensorPayload{
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: &temp,
		Humidity:    &humidity,
		Timestamp:   "2026-03-29T12:00:00Z",
	}
	reading := p.ToReading()
	if reading.SensorID != "esp32-001" {
		t.Errorf("expected sensor_id esp32-001, got %s", reading.SensorID)
	}
	if reading.Temperature != 25.5 {
		t.Errorf("expected temperature 25.5, got %f", reading.Temperature)
	}
	if reading.Humidity != 78.0 {
		t.Errorf("expected humidity 78.0, got %f", reading.Humidity)
	}
	if reading.Time.IsZero() {
		t.Error("expected non-zero time")
	}
}

func TestSensorPayload_ToReading_NoTimestamp(t *testing.T) {
	temp := 25.0
	humidity := 75.0
	p := mqtt.SensorPayload{
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: &temp,
		Humidity:    &humidity,
	}
	reading := p.ToReading()
	if reading.Time.IsZero() {
		t.Error("expected current UTC time when no timestamp provided")
	}
}

func TestSubscriber_Topic(t *testing.T) {
	handler := func(_ domain.EnvironmentReading) error { return nil }
	sub := mqtt.NewSubscriber("urushi/sensors/+", handler)
	if sub.Topic() != "urushi/sensors/+" {
		t.Errorf("expected topic urushi/sensors/+, got %s", sub.Topic())
	}
}

func TestSubscriber_ParseMessage_Valid(t *testing.T) {
	var received domain.EnvironmentReading
	handler := func(r domain.EnvironmentReading) error {
		received = r
		return nil
	}
	sub := mqtt.NewSubscriber("urushi/sensors/+", handler)

	payload := `{"sensor_id":"esp32-001","location":"urushi_buro","temperature":25.0,"humidity":75.0}`
	if err := sub.ParseMessage([]byte(payload)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received.SensorID != "esp32-001" {
		t.Errorf("expected sensor_id esp32-001, got %s", received.SensorID)
	}
}

func TestSubscriber_ParseMessage_InvalidJSON(t *testing.T) {
	handler := func(_ domain.EnvironmentReading) error { return nil }
	sub := mqtt.NewSubscriber("urushi/sensors/+", handler)

	err := sub.ParseMessage([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON payload") {
		t.Errorf("expected 'invalid JSON payload' in error, got: %v", err)
	}
}

func TestSubscriber_ParseMessage_ValidationFailure(t *testing.T) {
	handler := func(_ domain.EnvironmentReading) error { return nil }
	sub := mqtt.NewSubscriber("urushi/sensors/+", handler)

	// Missing required fields
	payload := `{"sensor_id":"","location":"urushi_buro","temperature":25.0,"humidity":75.0}`
	err := sub.ParseMessage([]byte(payload))
	if err == nil {
		t.Error("expected error for empty sensor_id")
	}
	if !strings.Contains(err.Error(), "payload validation failed") {
		t.Errorf("expected 'payload validation failed' in error, got: %v", err)
	}
}

func TestSubscriber_ParseMessage_MissingTemperature(t *testing.T) {
	handler := func(_ domain.EnvironmentReading) error { return nil }
	sub := mqtt.NewSubscriber("urushi/sensors/+", handler)

	payload := `{"sensor_id":"esp32-001","location":"urushi_buro","humidity":75.0}`
	err := sub.ParseMessage([]byte(payload))
	if err == nil {
		t.Error("expected error for missing temperature")
	}
}

func TestSubscriber_ParseMessage_BoundaryValues(t *testing.T) {
	handler := func(_ domain.EnvironmentReading) error { return nil }
	sub := mqtt.NewSubscriber("urushi/sensors/+", handler)

	tests := []struct {
		name    string
		payload string
		wantErr bool
	}{
		{
			"min temp boundary",
			`{"sensor_id":"esp32-001","location":"urushi_buro","temperature":-10.0,"humidity":50.0}`,
			false,
		},
		{
			"max temp boundary",
			`{"sensor_id":"esp32-001","location":"urushi_buro","temperature":100.0,"humidity":50.0}`,
			false,
		},
		{
			"min humidity boundary",
			`{"sensor_id":"esp32-001","location":"urushi_buro","temperature":25.0,"humidity":0.0}`,
			false,
		},
		{
			"max humidity boundary",
			`{"sensor_id":"esp32-001","location":"urushi_buro","temperature":25.0,"humidity":100.0}`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sub.ParseMessage([]byte(tt.payload))
			if (err != nil) != tt.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
