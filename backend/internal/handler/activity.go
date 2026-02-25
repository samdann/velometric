package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/velometric/backend/internal/fitparser"
)

// UploadResponse is returned after successful FIT file upload
type UploadResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ErrorResponse is returned on errors
type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	// TODO: Get user ID from auth context
	// For now, use a hardcoded demo user ID
	userID, err := getDemoUserID(r.Context(), h.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	activities, err := h.activityService.ListActivities(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list activities")
		return
	}

	writeJSON(w, http.StatusOK, activities)
}

func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}

	activity, err := h.activityService.GetActivity(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get activity")
		return
	}

	if activity == nil {
		writeError(w, http.StatusNotFound, "Activity not found")
		return
	}

	writeJSON(w, http.StatusOK, activity)
}

func (h *Handler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get the FIT file
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Parse the FIT file
	parsed, err := fitparser.Parse(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse FIT file: "+err.Error())
		return
	}

	// TODO: Get user ID and FTP from auth context
	userID, err := getDemoUserID(r.Context(), h.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	ftp := 250 // TODO: Get from user profile

	// Process and store the activity
	activity, err := h.activityService.ProcessFITFile(r.Context(), userID, parsed, ftp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to process activity: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, UploadResponse{
		ID:      activity.ID.String(),
		Message: "Activity uploaded successfully",
	})
}

// GetActivityRecords returns time-series data for an activity
func (h *Handler) GetActivityRecords(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}

	records, err := h.activityService.GetActivityRecords(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get records")
		return
	}

	writeJSON(w, http.StatusOK, records)
}

// GetPowerCurve returns the power curve for an activity
func (h *Handler) GetPowerCurve(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}

	curve, err := h.activityService.GetPowerCurve(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get power curve")
		return
	}

	writeJSON(w, http.StatusOK, curve)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
