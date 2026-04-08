package alert_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/alert"
	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

// mockHTTPClient is a test double for alert.HTTPClient.
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// mockResolver is a test double for alert.DNSResolver.
type mockResolver struct {
	ips []string
	err error
}

func (m *mockResolver) LookupHost(_ string) ([]string, error) {
	return m.ips, m.err
}

// publicResolver returns a public IP for any hostname.
func publicResolver() *mockResolver {
	return &mockResolver{ips: []string{"93.184.215.14"}}
}

func testReading() domain.EnvironmentReading {
	return domain.EnvironmentReading{
		Time:        time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 35.0,
		Humidity:    90.0,
	}
}

func testThreshold() domain.AlertThreshold {
	return domain.AlertThreshold{
		ID:             uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
	}
}

func TestWebhookNotifier_Notify_Success(t *testing.T) {
	var capturedBody []byte
	var capturedContentType string

	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			capturedContentType = req.Header.Get("Content-Type")
			body, _ := io.ReadAll(req.Body)
			capturedBody = body
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		},
	}

	notifier, err := alert.NewWebhookNotifierWithResolver("https://hooks.example.com/alert", client, publicResolver())
	if err != nil {
		t.Fatalf("NewWebhookNotifierWithResolver: %v", err)
	}
	err = notifier.Notify(testReading(), testThreshold())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", capturedContentType)
	}

	var payload alert.WebhookPayload
	if err := json.Unmarshal(capturedBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if payload.AlertType != "environment_threshold_exceeded" {
		t.Errorf("expected alert_type environment_threshold_exceeded, got %s", payload.AlertType)
	}
	if payload.SensorID != "esp32-001" {
		t.Errorf("expected sensor_id esp32-001, got %s", payload.SensorID)
	}
	if payload.Temperature != 35.0 {
		t.Errorf("expected temperature 35.0, got %f", payload.Temperature)
	}
	if !strings.Contains(payload.Message, "temperature") {
		t.Errorf("expected message to contain 'temperature', got: %s", payload.Message)
	}
	if !strings.Contains(payload.Message, "humidity") {
		t.Errorf("expected message to contain 'humidity', got: %s", payload.Message)
	}
}

func TestWebhookNotifier_Notify_EmptyURL_NoOp(t *testing.T) {
	notifier, err := alert.NewWebhookNotifier("", nil)
	if err != nil {
		t.Fatalf("NewWebhookNotifier: %v", err)
	}
	err = notifier.Notify(testReading(), testThreshold())
	if err != nil {
		t.Fatalf("expected no error for empty webhook URL, got %v", err)
	}
}

func TestWebhookNotifier_Notify_NonSuccessStatus(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		},
	}

	notifier, err := alert.NewWebhookNotifierWithResolver("https://hooks.example.com/alert", client, publicResolver())
	if err != nil {
		t.Fatalf("NewWebhookNotifierWithResolver: %v", err)
	}
	err = notifier.Notify(testReading(), testThreshold())
	if err == nil {
		t.Error("expected error for non-success status code")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error to mention status 500, got: %v", err)
	}
}

func TestWebhookNotifier_Notify_NetworkError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return nil, io.ErrUnexpectedEOF
		},
	}

	notifier, err := alert.NewWebhookNotifierWithResolver("https://hooks.example.com/alert", client, publicResolver())
	if err != nil {
		t.Fatalf("NewWebhookNotifierWithResolver: %v", err)
	}
	err = notifier.Notify(testReading(), testThreshold())
	if err == nil {
		t.Error("expected error for network failure")
	}
	if !strings.Contains(err.Error(), "failed to send alert webhook") {
		t.Errorf("expected 'failed to send alert webhook' in error, got: %v", err)
	}
}

func TestWebhookNotifier_AlertMessage_TemperatureLow(t *testing.T) {
	var capturedBody []byte
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			capturedBody = body
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		},
	}

	notifier, err := alert.NewWebhookNotifierWithResolver("https://hooks.example.com/alert", client, publicResolver())
	if err != nil {
		t.Fatalf("NewWebhookNotifierWithResolver: %v", err)
	}
	reading := testReading()
	reading.Temperature = 15.0
	reading.Humidity = 75.0 // within range

	err = notifier.Notify(reading, testThreshold())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var payload alert.WebhookPayload
	if err := json.Unmarshal(capturedBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !strings.Contains(payload.Message, "below minimum") {
		t.Errorf("expected message to contain 'below minimum', got: %s", payload.Message)
	}
}

func TestNewWebhookNotifier_RejectsInvalidScheme(t *testing.T) {
	_, err := alert.NewWebhookNotifierWithResolver("ftp://example.com/hook", nil, publicResolver())
	if err == nil {
		t.Fatal("expected error for ftp scheme, got nil")
	}
	if !strings.Contains(err.Error(), "scheme") {
		t.Errorf("expected error about scheme, got: %v", err)
	}
}

func TestNewWebhookNotifier_RejectsPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{name: "loopback", ip: "127.0.0.1"},
		{name: "private 10.x", ip: "10.0.0.1"},
		{name: "private 192.168.x", ip: "192.168.1.1"},
		{name: "private 172.16.x", ip: "172.16.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &mockResolver{ips: []string{tt.ip}}
			_, err := alert.NewWebhookNotifierWithResolver("https://evil.example.com/hook", nil, resolver)
			if err == nil {
				t.Fatalf("expected error for private IP %s, got nil", tt.ip)
			}
			if !strings.Contains(err.Error(), "private IP") {
				t.Errorf("expected 'private IP' in error, got: %v", err)
			}
		})
	}
}

func TestNewWebhookNotifier_AcceptsPublicIP(t *testing.T) {
	resolver := &mockResolver{ips: []string{"93.184.215.14"}}
	notifier, err := alert.NewWebhookNotifierWithResolver("https://hooks.example.com/alert", nil, resolver)
	if err != nil {
		t.Fatalf("expected no error for public IP, got %v", err)
	}
	if notifier == nil {
		t.Fatal("expected non-nil notifier")
	}
}
