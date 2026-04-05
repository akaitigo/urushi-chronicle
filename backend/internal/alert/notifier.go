// Package alert provides webhook notification for environment alert events.
package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
)

// Notifier sends alert notifications via webhook.
type Notifier interface {
	// Notify sends an alert notification for the given reading and threshold violation.
	Notify(reading domain.EnvironmentReading, threshold domain.AlertThreshold) error
}

// WebhookPayload is the JSON body sent to the webhook endpoint.
type WebhookPayload struct {
	AlertType   string    `json:"alert_type"`
	SensorID    string    `json:"sensor_id"`
	Location    string    `json:"location"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	ThresholdID string    `json:"threshold_id"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// WebhookNotifier sends alert notifications to a configurable webhook URL.
type WebhookNotifier struct {
	webhookURL string
	client     HTTPClient
}

// HTTPClient is an interface for sending HTTP requests, allowing test doubles.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewWebhookNotifier creates a new WebhookNotifier.
// If webhookURL is empty, notifications are silently dropped (no-op mode).
func NewWebhookNotifier(webhookURL string, client HTTPClient) *WebhookNotifier {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &WebhookNotifier{
		webhookURL: webhookURL,
		client:     client,
	}
}

// Notify sends an alert notification to the configured webhook URL.
// If webhookURL is empty, the notification is silently dropped.
func (n *WebhookNotifier) Notify(reading domain.EnvironmentReading, threshold domain.AlertThreshold) error {
	if n.webhookURL == "" {
		return nil
	}

	message := buildAlertMessage(reading, threshold)

	payload := WebhookPayload{
		AlertType:   "environment_threshold_exceeded",
		SensorID:    reading.SensorID,
		Location:    reading.Location,
		Temperature: reading.Temperature,
		Humidity:    reading.Humidity,
		ThresholdID: threshold.ID.String(),
		Message:     message,
		Timestamp:   reading.Time,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal alert payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create alert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send alert webhook: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // best-effort close; response already consumed
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func buildAlertMessage(reading domain.EnvironmentReading, threshold domain.AlertThreshold) string {
	var violations []string
	if reading.Temperature < threshold.TemperatureMin {
		violations = append(violations, fmt.Sprintf(
			"temperature %.1f°C is below minimum %.1f°C",
			reading.Temperature, threshold.TemperatureMin,
		))
	}
	if reading.Temperature > threshold.TemperatureMax {
		violations = append(violations, fmt.Sprintf(
			"temperature %.1f°C exceeds maximum %.1f°C",
			reading.Temperature, threshold.TemperatureMax,
		))
	}
	if reading.Humidity < threshold.HumidityMin {
		violations = append(violations, fmt.Sprintf(
			"humidity %.1f%% is below minimum %.1f%%",
			reading.Humidity, threshold.HumidityMin,
		))
	}
	if reading.Humidity > threshold.HumidityMax {
		violations = append(violations, fmt.Sprintf(
			"humidity %.1f%% exceeds maximum %.1f%%",
			reading.Humidity, threshold.HumidityMax,
		))
	}

	msg := fmt.Sprintf("[urushi-chronicle] Alert from sensor %s at %s: ", reading.SensorID, reading.Location)
	for i, v := range violations {
		if i > 0 {
			msg += "; "
		}
		msg += v
	}
	return msg
}
