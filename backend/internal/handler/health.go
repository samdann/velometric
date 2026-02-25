package handler

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Database string `json:"database"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "disconnected"

	if h.HasDB() {
		if err := h.db.Health(r.Context()); err == nil {
			dbStatus = "connected"
		} else {
			dbStatus = "error"
		}
	}

	response := HealthResponse{
		Status:   "ok",
		Version:  "0.1.0",
		Database: dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
