package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/velometric/backend/internal/config"
)

// dummyHandler is an http.Handler that always responds 200 OK.
var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// newCORSHandler wraps dummyHandler with the CORS middleware for a given origin.
func newCORSHandler(allowedOrigin string) http.Handler {
	cfg := &config.Config{FrontendURL: allowedOrigin}
	return CORS(cfg)(dummyHandler)
}

// ── preflight (OPTIONS) ───────────────────────────────────────────────────────

func TestCORS_PreflightAllowedOrigin(t *testing.T) {
	const origin = "http://localhost:3001"
	handler := newCORSHandler(origin)

	req := httptest.NewRequest(http.MethodOptions, "/api/activities", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Origin")
	if got != origin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
}

func TestCORS_PreflightAllowsConfiguredMethods(t *testing.T) {
	const origin = "http://localhost:3001"
	handler := newCORSHandler(origin)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "DELETE")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	methods := rr.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("Access-Control-Allow-Methods header is empty on preflight")
	}
}

func TestCORS_PreflightAllowsConfiguredHeaders(t *testing.T) {
	const origin = "http://localhost:3001"
	handler := newCORSHandler(origin)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	allowedHeaders := rr.Header().Get("Access-Control-Allow-Headers")
	if allowedHeaders == "" {
		t.Error("Access-Control-Allow-Headers header is empty on preflight")
	}
}

func TestCORS_PreflightAllowCredentials(t *testing.T) {
	const origin = "http://localhost:3001"
	handler := newCORSHandler(origin)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "GET")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	cred := rr.Header().Get("Access-Control-Allow-Credentials")
	if cred != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", cred, "true")
	}
}

// ── simple (non-preflight) requests ──────────────────────────────────────────

func TestCORS_SimpleRequestAllowedOrigin(t *testing.T) {
	const origin = "http://localhost:3001"
	handler := newCORSHandler(origin)

	req := httptest.NewRequest(http.MethodGet, "/api/activities", nil)
	req.Header.Set("Origin", origin)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Origin")
	if got != origin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestCORS_DisallowedOriginGetsNoAllowHeader(t *testing.T) {
	handler := newCORSHandler("http://localhost:3001")

	req := httptest.NewRequest(http.MethodGet, "/api/activities", nil)
	req.Header.Set("Origin", "https://evil.example.com")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// chi/cors responds with an empty ACAO header for disallowed origins
	got := rr.Header().Get("Access-Control-Allow-Origin")
	if got == "https://evil.example.com" {
		t.Errorf("disallowed origin should not be reflected in Access-Control-Allow-Origin")
	}
}

func TestCORS_NoOriginHeaderPassesThrough(t *testing.T) {
	// Requests without Origin header (same-origin, curl, healthchecks) must not be blocked.
	handler := newCORSHandler("http://localhost:3001")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("request without Origin header: status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// ── configuration is respected ───────────────────────────────────────────────

func TestCORS_FrontendURLIsUsedAsAllowedOrigin(t *testing.T) {
	// Verify that the FrontendURL in config is actually the allowed origin.
	const frontendURL = "https://velometric.io"
	handler := newCORSHandler(frontendURL)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", frontendURL)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Origin")
	if got != frontendURL {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, frontendURL)
	}
}
