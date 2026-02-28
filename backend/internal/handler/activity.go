package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/velometric/backend/internal/fitparser"
	"github.com/velometric/backend/internal/repository"
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
	log.Println("CreateActivity: starting")

	if !h.HasDB() {
		log.Println("CreateActivity: no database")
		writeError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("CreateActivity: form parse error: %v", err)
		writeError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	// Get the FIT file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("CreateActivity: no file: %v", err)
		writeError(w, http.StatusBadRequest, "No file provided: "+err.Error())
		return
	}
	defer file.Close()
	log.Printf("CreateActivity: received file %s (%d bytes)", header.Filename, header.Size)

	// Parse the FIT file
	parsed, err := fitparser.Parse(file)
	if err != nil {
		log.Printf("CreateActivity: FIT parse error: %v", err)
		writeError(w, http.StatusBadRequest, "Failed to parse FIT file: "+err.Error())
		return
	}
	log.Printf("CreateActivity: parsed %d records, %d laps", len(parsed.Records), len(parsed.Laps))

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
		if errors.Is(err, repository.ErrDuplicateActivity) {
			writeError(w, http.StatusConflict, "This activity has already been uploaded")
			return
		}
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

// GetElevationProfile returns the elevation profile for an activity
func (h *Handler) GetElevationProfile(w http.ResponseWriter, r *http.Request) {
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
	points, err := h.activityService.GetElevationProfile(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get elevation profile")
		return
	}
	writeJSON(w, http.StatusOK, points)
}

// GetSpeedProfile returns the speed profile for an activity
func (h *Handler) GetSpeedProfile(w http.ResponseWriter, r *http.Request) {
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
	points, err := h.activityService.GetSpeedProfile(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get speed profile")
		return
	}
	writeJSON(w, http.StatusOK, points)
}

// GetLaps returns the laps for an activity
func (h *Handler) GetLaps(w http.ResponseWriter, r *http.Request) {
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

	laps, err := h.activityService.GetLaps(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get laps")
		return
	}

	writeJSON(w, http.StatusOK, laps)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("writeJSON marshal error: %v", err)
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
