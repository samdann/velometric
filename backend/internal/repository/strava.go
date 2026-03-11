package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/velometric/backend/internal/model"
)

type StravaRepository struct {
	pool *pgxpool.Pool
}

func NewStravaRepository(pool *pgxpool.Pool) *StravaRepository {
	return &StravaRepository{pool: pool}
}

// Upsert inserts or updates a Strava activity
func (r *StravaRepository) Upsert(ctx context.Context, s *model.StravaActivity) error {
	rawJSON, err := json.Marshal(s.RawJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal raw_json: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO strava_activities (
			id, user_id, strava_id, title, activity_type, start_time, distance,
			is_private, is_flagged, raw_json, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id, strava_id) DO UPDATE SET
			title = EXCLUDED.title,
			activity_type = EXCLUDED.activity_type,
			start_time = EXCLUDED.start_time,
			distance = EXCLUDED.distance,
			is_private = EXCLUDED.is_private,
			is_flagged = EXCLUDED.is_flagged,
			raw_json = EXCLUDED.raw_json,
			synced_at = EXCLUDED.synced_at
	`,
		s.ID, s.UserID, s.StravaID, s.Title, s.ActivityType, s.StartTime, s.Distance,
		s.IsPrivate, s.IsFlagged, rawJSON, s.SyncedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert strava_activity: %w", err)
	}
	return nil
}

// GetByUserID retrieves all Strava activities for a user
func (r *StravaRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.StravaActivity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, strava_id, title, activity_type, start_time, distance,
			is_private, is_flagged, raw_json, synced_at
		FROM strava_activities
		WHERE user_id = $1
		ORDER BY start_time DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query strava_activities: %w", err)
	}
	defer rows.Close()

	var activities []*model.StravaActivity
	for rows.Next() {
		var a model.StravaActivity
		var rawJSON []byte
		err := rows.Scan(
			&a.ID, &a.UserID, &a.StravaID, &a.Title, &a.ActivityType, &a.StartTime, &a.Distance,
			&a.IsPrivate, &a.IsFlagged, &rawJSON, &a.SyncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan strava_activity: %w", err)
		}
		if rawJSON != nil {
			json.Unmarshal(rawJSON, &a.RawJSON)
		}
		activities = append(activities, &a)
	}
	return activities, nil
}

// FindByTimeRange finds local activities within a time window (for matching)
func (r *ActivityRepository) FindByTimeRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*model.Activity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, sport, start_time, duration, distance, elevation_gain,
			avg_power, max_power, normalized_power, tss, intensity_factor, variability_index,
			avg_hr, max_hr, avg_cadence, max_cadence, avg_speed, max_speed,
			calories, avg_temperature, fit_file_url, device_name, location,
			created_at, updated_at
		FROM activities
		WHERE user_id = $1 AND start_time >= $2 AND start_time <= $3
		ORDER BY start_time
	`, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities by time range: %w", err)
	}
	defer rows.Close()

	var activities []*model.Activity
	for rows.Next() {
		var a model.Activity
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Name, &a.Sport, &a.StartTime, &a.Duration, &a.Distance, &a.ElevationGain,
			&a.AvgPower, &a.MaxPower, &a.NormalizedPower, &a.TSS, &a.IntensityFactor, &a.VariabilityIndex,
			&a.AvgHR, &a.MaxHR, &a.AvgCadence, &a.MaxCadence, &a.AvgSpeed, &a.MaxSpeed,
			&a.Calories, &a.AvgTemperature, &a.FitFileURL, &a.DeviceName, &a.Location,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, &a)
	}
	return activities, nil
}

// UpdateActivity updates only specific fields of an activity
func (r *ActivityRepository) UpdateActivity(ctx context.Context, id uuid.UUID, name, sport string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE activities
		SET name = $1, sport = $2, updated_at = NOW()
		WHERE id = $3
	`, name, sport, id)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	return nil
}
