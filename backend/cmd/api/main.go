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
	"github.com/joho/godotenv"

	"github.com/velometric/backend/internal/config"
	"github.com/velometric/backend/internal/database"
	"github.com/velometric/backend/internal/handler"
	"github.com/velometric/backend/internal/middleware"
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
	h := handler.New(db)

	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(middleware.CORS(cfg))

	// Routes
	r.Get("/health", h.Health)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/activities", h.ListActivities)
		r.Post("/activities", h.CreateActivity)
		r.Get("/activities/{id}", h.GetActivity)
	})

	// Server setup
	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
