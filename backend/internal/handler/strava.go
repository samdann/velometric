package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/velometric/backend/internal/service"
)

// StravaHandler handles Strava-related HTTP requests
type StravaHandler struct {
	stravaService *service.StravaService
	getUserID     func(ctx context.Context) (uuid.UUID, error)
}

// NewStravaHandler creates a new Strava handler
func NewStravaHandler(stravaService *service.StravaService, getUserID func(ctx context.Context) (uuid.UUID, error)) *StravaHandler {
	return &StravaHandler{stravaService: stravaService, getUserID: getUserID}
}

// Sync creates an async Strava sync job and returns it immediately.
// @Summary Start async Strava sync
// @Description Creates a sync job and starts processing in background
// @Tags strava
// @Produce json
// @Param limit query int false "Max number of activities to process (0 = all)"
// @Success 202 {object} model.StravaSyncJob
// @Router /api/strava/sync [post]
func (h *StravaHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}

	limit := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	job, err := h.stravaService.StartSync(r.Context(), userID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

// GetJob returns the current state of a sync job.
// @Summary Get sync job status
// @Tags strava
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} model.StravaSyncJob
// @Router /api/strava/jobs/{id} [get]
func (h *StravaHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid job ID"})
		return
	}

	job, err := h.stravaService.GetJob(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// RetryJob retries a failed sync job from its failure point.
// @Summary Retry failed sync job
// @Tags strava
// @Produce json
// @Param id path string true "Job ID"
// @Success 202 {object} model.StravaSyncJob
// @Router /api/strava/jobs/{id}/retry [post]
func (h *StravaHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid job ID"})
		return
	}

	job, err := h.stravaService.RetrySync(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrJobNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		// Not in a failed state
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

// GetStatus returns the status of Strava integration
// @Summary Get Strava integration status
// @Tags strava
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /api/strava/status [get]
func (h *StravaHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{
		"configured": h.stravaService.HasToken(),
	})
}
