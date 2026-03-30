package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/handler"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/akaitigo/urushi-chronicle/internal/storage"
	"github.com/google/uuid"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Initialize repositories
	workRepo := repository.NewMemoryWorkRepository()
	stepRepo := repository.NewMemoryStepRepository()

	// Seed a sample work for MVP demo
	sampleWork := &domain.Work{
		ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Title:     "棗（なつめ）蒔絵",
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	workRepo.Seed(sampleWork)

	// Initialize storage
	gcsBucket := os.Getenv("GCS_BUCKET")
	uploader := storage.NewGCSUploader(gcsBucket)

	// Initialize handlers
	stepHandler := handler.NewStepHandler(stepRepo, workRepo, uploader)
	templateHandler := handler.NewTemplateHandler()

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthHandler())
	mux.Handle("/api/v1/works/", stepHandler)
	mux.Handle("/api/v1/templates", templateHandler)
	mux.Handle("/api/v1/templates/", templateHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Printf("urushi-chronicle API server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatalf("server failed: %v", err)
	}
}
