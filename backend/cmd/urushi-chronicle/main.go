package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/alert"
	"github.com/akaitigo/urushi-chronicle/internal/database"
	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/handler"
	"github.com/akaitigo/urushi-chronicle/internal/monitor"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/akaitigo/urushi-chronicle/internal/storage"
	"github.com/akaitigo/urushi-chronicle/pkg/middleware"
	"github.com/google/uuid"
)

// validateStoreType checks that storeType is one of the allowed values.
// Returns an error for invalid values.
func validateStoreType(storeType string) error {
	switch storeType {
	case "postgres", "memory", "":
		return nil
	default:
		return fmt.Errorf("invalid STORE_TYPE=%q: must be \"postgres\", \"memory\", or empty", storeType)
	}
}

func main() {
	logger := log.New(os.Stdout, "[urushi-chronicle] ", log.LstdFlags)

	// Initialize repositories based on STORE_TYPE environment variable.
	// STORE_TYPE=postgres  → use PostgreSQL (requires DATABASE_URL)
	// STORE_TYPE=memory    → use in-memory stores (default)
	//
	// For backward compatibility, if STORE_TYPE is not set but DATABASE_URL is present,
	// PostgreSQL is used automatically.
	var envRepo repository.EnvironmentRepository
	var thresholdRepo repository.AlertThresholdRepository
	var workRepo repository.WorkRepository
	var stepRepo repository.StepRepository

	storeType := os.Getenv("STORE_TYPE")
	databaseURL := os.Getenv("DATABASE_URL")

	// Validate STORE_TYPE before proceeding.
	if err := validateStoreType(storeType); err != nil {
		logger.Fatalf("%v", err)
	}

	// Determine effective store type
	usePostgres := storeType == "postgres" || (storeType == "" && databaseURL != "")

	if usePostgres {
		if databaseURL == "" {
			logger.Fatal("STORE_TYPE=postgres requires DATABASE_URL to be set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), database.DefaultConnectTimeout)
		defer cancel()

		pool, err := database.NewPool(ctx, databaseURL)
		if err != nil {
			logger.Fatalf("failed to connect to database: %v", err)
		}
		defer pool.Close()

		envRepo = repository.NewPgEnvironmentRepository(pool)
		thresholdRepo = repository.NewPgAlertThresholdRepository(pool)
		workRepo = repository.NewPgWorkRepository(pool)
		stepRepo = repository.NewPgStepRepository(pool)

		logger.Println("database: connected to PostgreSQL (STORE_TYPE=postgres)")
	} else {
		logger.Println("database: using in-memory stores (STORE_TYPE=memory)")

		memEnvRepo := repository.NewMemoryEnvironmentRepository()
		memThresholdRepo := repository.NewMemoryAlertThresholdRepository()
		memWorkRepo := repository.NewMemoryWorkRepository()
		memStepRepo := repository.NewMemoryStepRepository()

		// Seed demo data for in-memory mode
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
		memThresholdRepo.Seed(defaultThreshold)

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
		memWorkRepo.Seed(demoWork)

		envRepo = memEnvRepo
		thresholdRepo = memThresholdRepo
		workRepo = memWorkRepo
		stepRepo = memStepRepo
	}

	// Initialize alert notifier (webhook URL from env; empty = no-op)
	webhookURL := os.Getenv("ALERT_WEBHOOK_URL")
	notifier, err := alert.NewWebhookNotifier(webhookURL, nil)
	if err != nil {
		logger.Fatalf("failed to initialize alert notifier: %v", err)
	}

	// Initialize monitoring service
	monitorSvc := monitor.NewService(envRepo, thresholdRepo, notifier, logger)

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
