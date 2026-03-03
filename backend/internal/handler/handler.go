package handler

import (
	"github.com/velometric/backend/internal/database"
	"github.com/velometric/backend/internal/repository"
	"github.com/velometric/backend/internal/service"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db              *database.DB
	activityService *service.ActivityService
	userService     *service.UserService
}

// New creates a new Handler with dependencies
func New(db *database.DB) *Handler {
	h := &Handler{
		db: db,
	}

	if db != nil {
		activityRepo := repository.NewActivityRepository(db.Pool)
		h.activityService = service.NewActivityService(activityRepo)
		userRepo := repository.NewUserRepository(db.Pool)
		h.userService = service.NewUserService(userRepo)
	}

	return h
}

// HasDB returns true if database is available
func (h *Handler) HasDB() bool {
	return h.db != nil
}
