package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/velometric/backend/internal/config"
	"github.com/velometric/backend/internal/database"
	"github.com/velometric/backend/internal/handler"
	"github.com/velometric/backend/internal/middleware"
	"github.com/velometric/backend/internal/service"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := config.Load()

	// Initialize database (optional - app can run without it for basic testing)
	var db *database.DB
	ctx := context.Background()

	dbConfig := database.DefaultConfig(cfg.DatabaseURL)
	var err error
	db, err = database.New(ctx, dbConfig)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Printf("Running without database - some features will be unavailable")
	} else {
		log.Printf("Database connected successfully")
		defer db.Close()
	}

	// Create handlers with dependencies
	h := handler.New(db, cfg)

	// Strava service and handler
	var stravaHandler *handler.StravaHandler
	if db != nil {
		stravaService := service.NewStravaService(cfg, db.Pool)
		stravaHandler = handler.NewStravaHandler(stravaService, func(ctx context.Context) (uuid.UUID, error) {
			return handler.GetDemoUserID(ctx, db)
		})
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(middleware.CORS(cfg))

	// Routes
	r.Get("/health", h.Health)
	r.Get("/docs", h.SwaggerUI)
	r.Get("/docs/openapi.yaml", h.OpenAPISpec)
	r.Get("/favicon.ico", h.Favicon)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/feed", h.GetFeed)
		r.Get("/activities", h.ListActivities)
		r.Post("/activities", h.CreateActivity)
		r.Get("/activities/{id}", h.GetActivity)
		r.Get("/activities/{id}/records", h.GetActivityRecords)
		r.Get("/activities/{id}/power-curve", h.GetPowerCurve)
		r.Get("/activities/{id}/elevation", h.GetElevationProfile)
		r.Get("/activities/{id}/speed", h.GetSpeedProfile)
		r.Get("/activities/{id}/hr-cadence", h.GetHRCadenceProfile)
		r.Get("/activities/{id}/laps", h.GetLaps)
		r.Get("/activities/{id}/route", h.GetRoute)
		r.Get("/activities/{id}/hr-zone-distribution", h.GetHRZoneDistribution)
		r.Get("/activities/{id}/power-zone-distribution", h.GetPowerZoneDistribution)
		r.Delete("/activities/{id}", h.DeleteActivity)

		r.Post("/internal/batch-import", h.StartBatchImport)
		r.Get("/internal/batch-import/{id}", h.GetBatchImportStatus)

		r.Get("/user/profile", h.GetProfile)
		r.Put("/user/profile", h.UpdateProfile)
		r.Get("/user/hr-zones", h.GetHRZones)
		r.Put("/user/hr-zones", h.SaveHRZones)
		r.Get("/user/power-zones", h.GetPowerZones)
		r.Put("/user/power-zones", h.SavePowerZones)

		// Strava routes
		if stravaHandler != nil {
			r.Post("/strava/sync", stravaHandler.Sync)
			r.Get("/strava/status", stravaHandler.GetStatus)
			r.Get("/strava/jobs/{id}", stravaHandler.GetJob)
			r.Post("/strava/jobs/{id}/retry", stravaHandler.RetryJob)
		}
	})

	// Server setup
	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	log.Printf("Frontend URL: %s", cfg.FrontendURL)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Println("Server stopped")
}
