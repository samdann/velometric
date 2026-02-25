package handler

import (
	"github.com/velometric/backend/internal/database"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db *database.DB
}

// New creates a new Handler with dependencies
func New(db *database.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// HasDB returns true if database is available
func (h *Handler) HasDB() bool {
	return h.db != nil
}
