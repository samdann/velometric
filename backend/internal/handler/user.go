package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	u, err := h.userService.GetProfile(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	var req struct {
		Name   string   `json:"name"`
		Email  string   `json:"email"`
		Weight *float64 `json:"weight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "name and email are required")
		return
	}
	u, err := h.userService.UpdateProfile(r.Context(), req.Name, req.Email, req.Weight)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) GetHRZones(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	maxHR, zones, err := h.userService.GetHRZones(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load HR zones")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"max_hr": maxHR,
		"zones":  zones,
	})
}

func (h *Handler) SaveHRZones(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	var req struct {
		MaxHR      int   `json:"max_hr"`
		Boundaries []int `json:"boundaries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	zones, err := h.userService.SaveHRZones(r.Context(), req.MaxHR, req.Boundaries)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"max_hr": req.MaxHR,
		"zones":  zones,
	})
}

func (h *Handler) GetPowerZones(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	ftp, zones, err := h.userService.GetPowerZones(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load power zones")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ftp":   ftp,
		"zones": zones,
	})
}

func (h *Handler) SavePowerZones(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	var req struct {
		FTP        int   `json:"ftp"`
		Boundaries []int `json:"boundaries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	zones, err := h.userService.SavePowerZones(r.Context(), req.FTP, req.Boundaries)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ftp":   req.FTP,
		"zones": zones,
	})
}
