package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/velometric/backend/internal/config"
	"github.com/velometric/backend/internal/handler"
	"github.com/velometric/backend/internal/middleware"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := config.Load()

	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.CORS(cfg))

	// Routes
	r.Get("/health", handler.Health)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/activities", handler.ListActivities)
		r.Post("/activities", handler.CreateActivity)
		r.Get("/activities/{id}", handler.GetActivity)
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Frontend URL: %s", cfg.FrontendURL)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
