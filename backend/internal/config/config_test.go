package config

import (
	"os"
	"testing"
)

// ── getEnv (private) ──────────────────────────────────────────────────────────

func TestGetEnv_ReturnsValueWhenSet(t *testing.T) {
	const key = "VELOMETRIC_TEST_KEY"
	t.Setenv(key, "test-value")

	got := getEnv(key, "fallback")
	if got != "test-value" {
		t.Errorf("getEnv(%q) = %q, want %q", key, got, "test-value")
	}
}

func TestGetEnv_ReturnsFallbackWhenUnset(t *testing.T) {
	const key = "VELOMETRIC_DEFINITELY_NOT_SET_XYZ"
	os.Unsetenv(key)

	got := getEnv(key, "default-value")
	if got != "default-value" {
		t.Errorf("getEnv(%q) = %q, want fallback %q", key, got, "default-value")
	}
}

func TestGetEnv_EmptyEnvVarIsNotFallback(t *testing.T) {
	// An env var that is set but empty is a valid value — it should NOT fall back.
	const key = "VELOMETRIC_TEST_EMPTY"
	t.Setenv(key, "")

	got := getEnv(key, "fallback")
	if got != "" {
		t.Errorf("empty env var should return empty string, not fallback; got %q", got)
	}
}

// ── Load ──────────────────────────────────────────────────────────────────────

func TestLoad_NeverNil(t *testing.T) {
	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() returned nil")
	}
}

func TestLoad_EnvOverridesDefaults(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("FRONTEND_URL", "https://app.example.com")
	t.Setenv("STRAVA_ACCESS_TOKEN", "tok-abc123")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.FrontendURL != "https://app.example.com" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "https://app.example.com")
	}
	if cfg.StravaAccessToken != "tok-abc123" {
		t.Errorf("StravaAccessToken = %q, want %q", cfg.StravaAccessToken, "tok-abc123")
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	const key = "PORT"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) }) // ensure clean state after test

	cfg := Load()
	if cfg.Port != "8081" {
		t.Errorf("default Port = %q, want %q", cfg.Port, "8081")
	}
}

func TestLoad_DefaultFrontendURL(t *testing.T) {
	const key = "FRONTEND_URL"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	cfg := Load()
	if cfg.FrontendURL != "http://localhost:3001" {
		t.Errorf("default FrontendURL = %q, want %q", cfg.FrontendURL, "http://localhost:3001")
	}
}

func TestLoad_StravaTokenEmptyByDefault(t *testing.T) {
	const key = "STRAVA_ACCESS_TOKEN"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	cfg := Load()
	if cfg.StravaAccessToken != "" {
		t.Errorf("StravaAccessToken should default to empty, got %q", cfg.StravaAccessToken)
	}
}
