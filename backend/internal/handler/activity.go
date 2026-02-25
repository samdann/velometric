package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Activity struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Date          string  `json:"date"`
	Duration      int     `json:"duration"`
	Distance      float64 `json:"distance"`
	ElevationGain float64 `json:"elevationGain"`
	AveragePower  *int    `json:"averagePower,omitempty"`
	AverageHR     *int    `json:"averageHeartRate,omitempty"`
}

func ListActivities(w http.ResponseWriter, r *http.Request) {
	// Placeholder - will be replaced with database query
	activities := []Activity{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

func GetActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Placeholder - will be replaced with database query
	activity := Activity{
		ID:   id,
		Name: "Placeholder Activity",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}

func CreateActivity(w http.ResponseWriter, r *http.Request) {
	// Placeholder - will handle FIT file upload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Activity upload endpoint - not yet implemented",
	})
}
