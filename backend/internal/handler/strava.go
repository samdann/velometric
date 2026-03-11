package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/velometric/backend/internal/service"
)

// StravaHandler handles Strava-related HTTP requests
type StravaHandler struct {
	stravaService *service.StravaService
	getUserID      func(ctx context.Context) (uuid.UUID, error)
}

// NewStravaHandler creates a new Strava handler
func NewStravaHandler(stravaService *service.StravaService, getUserID func(ctx context.Context) (uuid.UUID, error)) *StravaHandler {
	return &StravaHandler{stravaService: stravaService, getUserID: getUserID}
}

// SyncResult is the response for sync operations
type SyncResult struct {
	UpdatedCount int                      `json:"updatedCount"`
	CreatedCount int                      `json:"createdCount"`
	Candidates   []MatchCandidateResponse `json:"candidates"`
	Error        string                   `json:"error,omitempty"`
}

// MatchCandidateResponse represents a potential match
type MatchCandidateResponse struct {
	StravaID        int64   `json:"stravaId"`
	StravaTitle     string  `json:"stravaTitle"`
	LocalID         string  `json:"localId,omitempty"`
	LocalTitle      string  `json:"localTitle,omitempty"`
	TimeDiffSecs    int64   `json:"timeDiffSecs"`
	DistanceDiffPct float64 `json:"distanceDiffPct"`
}

// Sync triggers a Strava sync for the current user
// @Summary Sync activities from Strava
// @Description Fetches activities from Strava and syncs them to local database
// @Tags strava
// @Produce json
// @Success 200 {object} SyncResult
// @Router /api/strava/sync [post]
func (h *StravaHandler) Sync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.getUserID(ctx)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}

	result, err := h.stravaService.FetchAndSync(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Convert to response format
	response := SyncResult{
		UpdatedCount: result.UpdatedCount,
		CreatedCount: result.CreatedCount,
		Candidates:   make([]MatchCandidateResponse, 0, len(result.Candidates)),
	}

	for _, c := range result.Candidates {
		resp := MatchCandidateResponse{
			TimeDiffSecs:    c.TimeDiffSecs,
			DistanceDiffPct: c.DistanceDiffPct,
		}
		if c.StravaActivity != nil {
			resp.StravaID = c.StravaActivity.StravaID
			if c.StravaActivity.Title != nil {
				resp.StravaTitle = *c.StravaActivity.Title
			}
		}
		if c.LocalActivity != nil {
			resp.LocalID = c.LocalActivity.ID.String()
			resp.LocalTitle = c.LocalActivity.Name
		}
		response.Candidates = append(response.Candidates, resp)
	}

	writeJSON(w, http.StatusOK, response)
}

// GetStatus returns the status of Strava integration
// @Summary Get Strava integration status
// @Description Returns whether Strava is configured
// @Tags strava
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /api/strava/status [get]
func (h *StravaHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{
		"configured": h.stravaService.HasToken(),
	})
}
