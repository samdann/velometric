package handler

import (
	"os"
	"path/filepath"

	"github.com/velometric/backend/internal/config"
	"github.com/velometric/backend/internal/database"
	"github.com/velometric/backend/internal/repository"
	"github.com/velometric/backend/internal/service"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db              *database.DB
	cfg             *config.Config
	activityService *service.ActivityService
	userService     *service.UserService
	batchImport     *service.BatchImportService
	stravaService   *service.StravaService
}

// New creates a new Handler with dependencies
func New(db *database.DB, cfg *config.Config) *Handler {
	h := &Handler{db: db, cfg: cfg}

	if db != nil {
		activityRepo := repository.NewActivityRepository(db.Pool)
		h.activityService = service.NewActivityService(activityRepo)
		userRepo := repository.NewUserRepository(db.Pool)
		h.userService = service.NewUserService(userRepo)

		fitDir := os.Getenv("FIT_DIR")
		if fitDir == "" {
			fitDir = filepath.Join("..", ".fit")
		}
		h.batchImport = service.NewBatchImportService(activityRepo, fitDir)

		// Strava service (optional - may not be configured)
		h.stravaService = service.NewStravaService(cfg, db.Pool)
	}

	return h
}

// HasDB returns true if database is available
func (h *Handler) HasDB() bool {
	return h.db != nil
}
