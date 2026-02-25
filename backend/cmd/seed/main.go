package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/velometric?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database, seeding data...")

	if err := seed(ctx, pool); err != nil {
		log.Fatalf("Seed failed: %v", err)
	}

	log.Println("Seed completed successfully!")
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	// Create a test user
	var userID string
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, name, ftp, max_hr, weight)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, "demo@velometric.app", "Demo Cyclist", 250, 185, 75.0).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	log.Printf("Created/updated user: %s", userID)

	// Create default power zones for the user
	powerZones := []struct {
		num   int
		name  string
		minP  float64
		maxP  *float64
		color string
	}{
		{1, "Active Recovery", 0, ptr(55.0), "#64748B"},
		{2, "Endurance", 55, ptr(75.0), "#3B82F6"},
		{3, "Tempo", 75, ptr(90.0), "#22C55E"},
		{4, "Threshold", 90, ptr(105.0), "#EAB308"},
		{5, "VO2max", 105, ptr(120.0), "#F97316"},
		{6, "Anaerobic", 120, ptr(150.0), "#EF4444"},
		{7, "Neuromuscular", 150, nil, "#DC2626"},
	}

	for _, z := range powerZones {
		_, err := pool.Exec(ctx, `
			INSERT INTO user_power_zones (user_id, zone_number, name, min_percentage, max_percentage, color)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, zone_number) DO UPDATE SET
				name = EXCLUDED.name,
				min_percentage = EXCLUDED.min_percentage,
				max_percentage = EXCLUDED.max_percentage,
				color = EXCLUDED.color
		`, userID, z.num, z.name, z.minP, z.maxP, z.color)
		if err != nil {
			return fmt.Errorf("failed to create power zone %d: %w", z.num, err)
		}
	}
	log.Printf("Created power zones for user")

	// Create default HR zones
	hrZones := []struct {
		num   int
		name  string
		minP  float64
		maxP  *float64
		color string
	}{
		{1, "Recovery", 50, ptr(60.0), "#64748B"},
		{2, "Aerobic", 60, ptr(70.0), "#3B82F6"},
		{3, "Tempo", 70, ptr(80.0), "#22C55E"},
		{4, "Threshold", 80, ptr(90.0), "#F97316"},
		{5, "Max", 90, nil, "#EF4444"},
	}

	for _, z := range hrZones {
		_, err := pool.Exec(ctx, `
			INSERT INTO user_hr_zones (user_id, zone_number, name, min_percentage, max_percentage, color)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, zone_number) DO UPDATE SET
				name = EXCLUDED.name,
				min_percentage = EXCLUDED.min_percentage,
				max_percentage = EXCLUDED.max_percentage,
				color = EXCLUDED.color
		`, userID, z.num, z.name, z.minP, z.maxP, z.color)
		if err != nil {
			return fmt.Errorf("failed to create HR zone %d: %w", z.num, err)
		}
	}
	log.Printf("Created HR zones for user")

	// Create sample activities
	activities := []struct {
		name          string
		startTime     time.Time
		duration      int
		distance      float64
		elevationGain float64
		avgPower      int
		np            int
		avgHR         int
		maxHR         int
		avgCadence    int
		avgSpeed      float64
	}{
		{
			name:          "Morning Tempo Ride",
			startTime:     time.Now().AddDate(0, 0, -1).Add(-2 * time.Hour),
			duration:      5400,  // 1.5 hours
			distance:      45000, // 45 km
			elevationGain: 450,
			avgPower:      195,
			np:            210,
			avgHR:         145,
			maxHR:         172,
			avgCadence:    88,
			avgSpeed:      8.33, // 30 km/h
		},
		{
			name:          "Weekend Long Ride",
			startTime:     time.Now().AddDate(0, 0, -3).Add(-4 * time.Hour),
			duration:      14400, // 4 hours
			distance:      100000,
			elevationGain: 1200,
			avgPower:      175,
			np:            190,
			avgHR:         138,
			maxHR:         168,
			avgCadence:    85,
			avgSpeed:      6.94, // 25 km/h
		},
		{
			name:          "Interval Session",
			startTime:     time.Now().AddDate(0, 0, -5).Add(-1 * time.Hour),
			duration:      3600,
			distance:      32000,
			elevationGain: 200,
			avgPower:      220,
			np:            265,
			avgHR:         155,
			maxHR:         185,
			avgCadence:    92,
			avgSpeed:      8.89,
		},
	}

	for _, a := range activities {
		var activityID string
		tss := float64(a.duration) * float64(a.np) * (float64(a.np) / 250.0) / 36000.0
		intensityFactor := float64(a.np) / 250.0
		variabilityIndex := float64(a.np) / float64(a.avgPower)

		err := pool.QueryRow(ctx, `
			INSERT INTO activities (
				user_id, name, sport, start_time, duration, distance, elevation_gain,
				avg_power, normalized_power, tss, intensity_factor, variability_index,
				avg_hr, max_hr, avg_cadence, avg_speed
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			RETURNING id
		`, userID, a.name, "cycling", a.startTime, a.duration, a.distance, a.elevationGain,
			a.avgPower, a.np, tss, intensityFactor, variabilityIndex,
			a.avgHR, a.maxHR, a.avgCadence, a.avgSpeed).Scan(&activityID)
		if err != nil {
			return fmt.Errorf("failed to create activity %s: %w", a.name, err)
		}
		log.Printf("Created activity: %s (%s)", a.name, activityID)

		// Create power curve for the activity
		powerCurve := []struct {
			duration int
			power    int
		}{
			{1, a.avgPower + 400},
			{5, a.avgPower + 300},
			{10, a.avgPower + 200},
			{30, a.avgPower + 150},
			{60, a.avgPower + 100},
			{300, a.avgPower + 50},
			{600, a.avgPower + 30},
			{1200, a.avgPower + 15},
			{3600, a.avgPower},
		}

		for _, pc := range powerCurve {
			if pc.duration <= a.duration {
				_, err := pool.Exec(ctx, `
					INSERT INTO activity_power_curve (activity_id, duration_seconds, best_power)
					VALUES ($1, $2, $3)
					ON CONFLICT (activity_id, duration_seconds) DO UPDATE SET best_power = EXCLUDED.best_power
				`, activityID, pc.duration, pc.power)
				if err != nil {
					return fmt.Errorf("failed to create power curve point: %w", err)
				}
			}
		}
	}

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
