package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/velometric/backend/internal/service"
)

// ── POST /api/internal/batch-import ──────────────────────────────────────────

type batchImportRequest struct {
	// Types filters by sport (case-insensitive). Empty = all.
	// E.g. ["cycling", "running", "swimming"]
	Types []string `json:"types"`
	// From / To in RFC3339. Omit for no bound.
	From *string `json:"from,omitempty"`
	To   *string `json:"to,omitempty"`
}

func (h *Handler) StartBatchImport(w http.ResponseWriter, r *http.Request) {
	if !h.HasDB() {
		writeError(w, http.StatusServiceUnavailable, "database not available")
		return
	}

	var req batchImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	importReq := service.ImportRequest{Types: req.Types}

	if req.From != nil {
		t, err := time.Parse(time.RFC3339, *req.From)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid 'from' (use RFC3339): "+err.Error())
			return
		}
		importReq.From = &t
	}
	if req.To != nil {
		t, err := time.Parse(time.RFC3339, *req.To)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid 'to' (use RFC3339): "+err.Error())
			return
		}
		importReq.To = &t
	}

	userID, err := h.resolveUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to resolve user")
		return
	}

	ftp := 250
	if h.userService != nil {
		if profile, err := h.userService.GetProfile(r.Context()); err == nil && profile.FTP != nil {
			ftp = *profile.FTP
		}
	}

	job, err := h.batchImport.StartImport(r.Context(), userID, ftp, importReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

// ── GET /api/internal/batch-import/{id} ──────────────────────────────────────

func (h *Handler) GetBatchImportStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	job, ok := h.batchImport.GetJob(id)
	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	writeJSON(w, http.StatusOK, job)
}
