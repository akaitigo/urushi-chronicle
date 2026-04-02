package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/alert"
	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/handler"
	"github.com/akaitigo/urushi-chronicle/internal/monitor"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/akaitigo/urushi-chronicle/internal/storage"
	"github.com/akaitigo/urushi-chronicle/pkg/middleware"
	"github.com/google/uuid"
)

func main() {
	logger := log.New(os.Stdout, "[urushi-chronicle] ", log.LstdFlags)

	// Initialize repositories
	envRepo := repository.NewMemoryEnvironmentRepository()
	thresholdRepo := repository.NewMemoryAlertThresholdRepository()
	workRepo := repository.NewMemoryWorkRepository()
	stepRepo := repository.NewMemoryStepRepository()

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

	// Seed a default demo work for the frontend
	demoWorkID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	demoWork := &domain.Work{
		ID:          demoWorkID,
		Title:       "蒔絵香合 — 秋草",
		Description: "秋草文様の蒔絵を施した香合。研出蒔絵技法を用いた習作。",
		Technique:   domain.TechniqueMakie,
		Material:    "欅",
		Status:      domain.WorkStatusInProgress,
		StartedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	workRepo.Seed(demoWork)

	// Initialize HTTP handlers
	envHandler := handler.NewEnvironmentHandler(envRepo, thresholdRepo, monitorSvc)
	bucketName := os.Getenv("GCS_BUCKET")
	uploader := storage.NewGCSUploader(bucketName)
	workHandler := handler.NewWorkHandler(workRepo, stepRepo)
	stepHandler := handler.NewStepHandler(stepRepo, workRepo, uploader)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthHandler())
	mux.Handle("/api/v1/environment/", envHandler)

	// /api/v1/works/ prefix: dispatch to StepHandler when path contains "/steps",
	// otherwise delegate to WorkHandler for single-work lookups.
	mux.HandleFunc("/api/v1/works/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/steps") {
			stepHandler.ServeHTTP(w, r)
		} else {
			workHandler.ServeHTTP(w, r)
		}
	})
	mux.Handle("/api/v1/works", workHandler) // exact match: works list

	// Wrap with CORS middleware (configured via CORS_ORIGINS env var)
	corsOrigins := os.Getenv("CORS_ORIGINS")
	var rootHandler http.Handler = mux
	rootHandler = middleware.CORS(rootHandler, corsOrigins)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      rootHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Printf("API server starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("server failed: %v", err)
	}
}
