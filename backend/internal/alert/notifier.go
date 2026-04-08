// Package alert provides webhook notification for environment alert events.
package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
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

// DNSResolver is an interface for resolving hostnames to IP addresses.
// This allows test doubles to avoid actual DNS lookups.
type DNSResolver interface {
	LookupHost(host string) ([]string, error)
}

// defaultResolver uses the standard library's net.LookupHost.
type defaultResolver struct{}

func (defaultResolver) LookupHost(host string) ([]string, error) {
	return net.LookupHost(host)
}

// NewWebhookNotifier creates a new WebhookNotifier.
// If webhookURL is empty, notifications are silently dropped (no-op mode).
// Returns an error if the URL scheme is not http/https or if the host resolves to a private IP.
func NewWebhookNotifier(webhookURL string, client HTTPClient) (*WebhookNotifier, error) {
	return NewWebhookNotifierWithResolver(webhookURL, client, defaultResolver{})
}

// NewWebhookNotifierWithResolver creates a new WebhookNotifier with a custom DNS resolver.
// Use this in tests to avoid actual DNS lookups.
func NewWebhookNotifierWithResolver(webhookURL string, client HTTPClient, resolver DNSResolver) (*WebhookNotifier, error) {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if webhookURL != "" {
		if err := validateWebhookURL(webhookURL, resolver); err != nil {
			return nil, fmt.Errorf("invalid webhook URL: %w", err)
		}
	}
	return &WebhookNotifier{
		webhookURL: webhookURL,
		client:     client,
	}, nil
}

// validateWebhookURL checks that the URL uses http or https scheme
// and does not resolve to a private/loopback IP address.
func validateWebhookURL(rawURL string, resolver DNSResolver) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Only allow http and https schemes.
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("scheme %q is not allowed; only http and https are permitted", parsed.Scheme)
	}

	// Resolve host to IP addresses and reject private/loopback ranges.
	hostname := parsed.Hostname()
	ips, err := resolver.LookupHost(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve host %q: %w", hostname, err)
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if isPrivateIP(ip) {
			return fmt.Errorf("host %q resolves to private IP %s", hostname, ipStr)
		}
	}

	return nil
}

// isPrivateIP checks whether an IP address is in a private, loopback,
// or link-local range.
func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
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
