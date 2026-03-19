package handler

import (
	"net/http"
	"strconv"
)

func (h *Handler) GetStatisticsYears(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}
	userID, err := h.resolveUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to resolve user")
		return
	}
	years, err := h.statisticsService.GetAvailablePowerYears(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get available years")
		return
	}
	if years == nil {
		years = make([]int, 0)
	}
	writeJSON(w, http.StatusOK, years)
}

func (h *Handler) GetStatisticsPower(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		writeError(w, http.StatusBadRequest, "year parameter required")
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		writeError(w, http.StatusBadRequest, "invalid year")
		return
	}

	userID, err := h.resolveUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to resolve user")
		return
	}

	ftp, zones, err := h.userService.GetPowerZones(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get power zones")
		return
	}

	stats, err := h.statisticsService.GetAnnualPowerStats(r.Context(), userID, year, ftp, zones)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to compute statistics")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
