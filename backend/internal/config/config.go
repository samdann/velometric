package config

import "os"

type Config struct {
	Port              string
	FrontendURL       string
	DatabaseURL       string
	RedisURL          string
	StravaAccessToken string
	StravaClientID    string
	StravaSecret      string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8081"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3001"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/velometric?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		StravaAccessToken: getEnv("STRAVA_ACCESS_TOKEN", ""),
		StravaClientID:    getEnv("STRAVA_CLIENT_ID", ""),
		StravaSecret:      getEnv("STRAVA_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
