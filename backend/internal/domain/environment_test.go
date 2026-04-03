package domain_test

import (
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

func validEnvironmentReading() domain.EnvironmentReading {
	return domain.EnvironmentReading{
		Time:        time.Now(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    75.0,
	}
}

func TestEnvironmentReadingValidate_Valid(t *testing.T) {
	er := validEnvironmentReading()
	if err := er.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestEnvironmentReadingValidate_EmptySensorID(t *testing.T) {
	er := validEnvironmentReading()
	er.SensorID = ""
	if err := er.Validate(); err == nil {
		t.Error("expected error for empty sensor_id")
	}
}

func TestEnvironmentReadingValidate_EmptyLocation(t *testing.T) {
	er := validEnvironmentReading()
	er.Location = ""
	if err := er.Validate(); err == nil {
		t.Error("expected error for empty location")
	}
}

func TestEnvironmentReadingValidate_TemperatureTooLow(t *testing.T) {
	er := validEnvironmentReading()
	er.Temperature = -11.0
	if err := er.Validate(); err == nil {
		t.Error("expected error for temperature below -10.0")
	}
}

func TestEnvironmentReadingValidate_TemperatureTooHigh(t *testing.T) {
	er := validEnvironmentReading()
	er.Temperature = 101.0
	if err := er.Validate(); err == nil {
		t.Error("expected error for temperature above 100.0")
	}
}

func TestEnvironmentReadingValidate_HumidityTooLow(t *testing.T) {
	er := validEnvironmentReading()
	er.Humidity = -1.0
	if err := er.Validate(); err == nil {
		t.Error("expected error for humidity below 0.0")
	}
}

func TestEnvironmentReadingValidate_HumidityTooHigh(t *testing.T) {
	er := validEnvironmentReading()
	er.Humidity = 101.0
	if err := er.Validate(); err == nil {
		t.Error("expected error for humidity above 100.0")
	}
}

func TestEnvironmentReadingValidate_ZeroTime(t *testing.T) {
	er := validEnvironmentReading()
	er.Time = time.Time{}
	if err := er.Validate(); err == nil {
		t.Error("expected error for zero time")
	}
}

func TestEnvironmentReadingValidate_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		humidity    float64
		wantErr     bool
	}{
		{"min temp", -10.0, 50.0, false},
		{"max temp", 100.0, 50.0, false},
		{"min humidity", 25.0, 0.0, false},
		{"max humidity", 25.0, 100.0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			er := validEnvironmentReading()
			er.Temperature = tt.temperature
			er.Humidity = tt.humidity
			err := er.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvironmentReadingValidate_SensorIDTooLong(t *testing.T) {
	er := validEnvironmentReading()
	er.SensorID = string(make([]byte, 257))
	if err := er.Validate(); err == nil {
		t.Error("expected error for sensor_id exceeding 256 characters")
	}
}

func TestEnvironmentReadingValidate_LocationTooLong(t *testing.T) {
	er := validEnvironmentReading()
	er.Location = string(make([]byte, 257))
	if err := er.Validate(); err == nil {
		t.Error("expected error for location exceeding 256 characters")
	}
}

func TestEnvironmentReadingValidate_ControlCharsStripped(t *testing.T) {
	er := validEnvironmentReading()
	er.SensorID = "esp32\x00-001\n"
	er.Location = "\turushi\x07_buro"
	if err := er.Validate(); err != nil {
		t.Errorf("expected valid after stripping control chars, got %v", err)
	}
	if er.SensorID != "esp32-001" {
		t.Errorf("expected sanitized sensor_id 'esp32-001', got %q", er.SensorID)
	}
	if er.Location != "urushi_buro" {
		t.Errorf("expected sanitized location 'urushi_buro', got %q", er.Location)
	}
}

func TestAlertThresholdValidate_Valid(t *testing.T) {
	at := domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
	}
	if err := at.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAlertThresholdValidate_MinGreaterThanMax(t *testing.T) {
	at := domain.AlertThreshold{
		SensorID:       "esp32-001",
		TemperatureMin: 30.0,
		TemperatureMax: 20.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
	}
	if err := at.Validate(); err == nil {
		t.Error("expected error when temperature_min >= temperature_max")
	}
}

func TestAlertThreshold_IsOutOfRange(t *testing.T) {
	at := domain.AlertThreshold{
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
	}

	tests := []struct {
		name     string
		reading  domain.EnvironmentReading
		expected bool
	}{
		{
			"within range",
			domain.EnvironmentReading{Temperature: 25.0, Humidity: 75.0},
			false,
		},
		{
			"temperature too low",
			domain.EnvironmentReading{Temperature: 15.0, Humidity: 75.0},
			true,
		},
		{
			"temperature too high",
			domain.EnvironmentReading{Temperature: 35.0, Humidity: 75.0},
			true,
		},
		{
			"humidity too low",
			domain.EnvironmentReading{Temperature: 25.0, Humidity: 60.0},
			true,
		},
		{
			"humidity too high",
			domain.EnvironmentReading{Temperature: 25.0, Humidity: 90.0},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := at.IsOutOfRange(tt.reading); got != tt.expected {
				t.Errorf("IsOutOfRange() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAlertThreshold_IsOutOfRange_Disabled(t *testing.T) {
	at := domain.AlertThreshold{
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        false,
	}
	reading := domain.EnvironmentReading{Temperature: 15.0, Humidity: 60.0}
	if at.IsOutOfRange(reading) {
		t.Error("disabled threshold should never report out of range")
	}
}
