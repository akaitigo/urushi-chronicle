package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/alert"
	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/handler"
	"github.com/akaitigo/urushi-chronicle/internal/monitor"
	"github.com/akaitigo/urushi-chronicle/internal/mqtt"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

func main() {
	logger := log.New(os.Stdout, "[urushi-chronicle] ", log.LstdFlags)

	// Initialize repositories
	envRepo := repository.NewMemoryEnvironmentRepository()
	thresholdRepo := repository.NewMemoryAlertThresholdRepository()

	// Initialize alert notifier (webhook URL from env; empty = no-op)
	webhookURL := os.Getenv("ALERT_WEBHOOK_URL")
	notifier := alert.NewWebhookNotifier(webhookURL, nil)

	// Initialize monitoring service
	monitorSvc := monitor.NewService(envRepo, thresholdRepo, notifier, logger)

	// Seed a default alert threshold for the demo urushi-buro sensor
	now := time.Now().UTC()
	defaultThreshold := &domain.AlertThreshold{
		ID:             uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	thresholdRepo.Seed(defaultThreshold)

	// Initialize MQTT subscriber (topic from env or default)
	mqttTopic := os.Getenv("MQTT_TOPIC")
	if mqttTopic == "" {
		mqttTopic = "urushi/sensors/+"
	}
	_ = mqtt.NewSubscriber(mqttTopic, monitorSvc.ProcessReading)

	// Initialize HTTP handlers
	envHandler := handler.NewEnvironmentHandler(envRepo, thresholdRepo, monitorSvc)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthHandler())
	mux.Handle("/api/v1/environment/", envHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Printf("API server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatalf("server failed: %v", err)
	}
}
