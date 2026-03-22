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

// Upsert inserts or updates a Strava activity, tagging it with the given job ID.
func (r *StravaRepository) Upsert(ctx context.Context, s *model.StravaActivity, jobID uuid.UUID) error {
	rawJSON, err := json.Marshal(s.RawJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal raw_json: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO strava_activities (
			id, user_id, strava_id, title, activity_type, start_time, distance,
			is_private, is_flagged, raw_json, synced_at, job_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (user_id, strava_id) DO UPDATE SET
			title = EXCLUDED.title,
			activity_type = EXCLUDED.activity_type,
			start_time = EXCLUDED.start_time,
			distance = EXCLUDED.distance,
			is_private = EXCLUDED.is_private,
			is_flagged = EXCLUDED.is_flagged,
			raw_json = EXCLUDED.raw_json,
			synced_at = EXCLUDED.synced_at,
			job_id = EXCLUDED.job_id
	`,
		s.ID, s.UserID, s.StravaID, s.Title, s.ActivityType, s.StartTime, s.Distance,
		s.IsPrivate, s.IsFlagged, rawJSON, s.SyncedAt, jobID,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert strava_activity: %w", err)
	}
	return nil
}

// CreateJob persists a new sync job.
func (r *StravaRepository) CreateJob(ctx context.Context, job *model.StravaSyncJob) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO strava_sync_jobs (id, user_id, status, local_only, activity_type, limit_count, offset_count, started_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, job.ID, job.UserID, job.Status, job.LocalOnly, job.ActivityType, job.LimitCount, job.OffsetCount, job.StartedAt, job.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create strava_sync_job: %w", err)
	}
	return nil
}

// GetJob retrieves a job by ID.
func (r *StravaRepository) GetJob(ctx context.Context, id uuid.UUID) (*model.StravaSyncJob, error) {
	var j model.StravaSyncJob
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, local_only, activity_type, limit_count, offset_count, fetched_count, updated_count, created_count,
		       error_message, started_at, fetched_at, completed_at, created_at
		FROM strava_sync_jobs WHERE id = $1
	`, id).Scan(
		&j.ID, &j.UserID, &j.Status, &j.LocalOnly, &j.ActivityType, &j.LimitCount, &j.OffsetCount, &j.FetchedCount, &j.UpdatedCount, &j.CreatedCount,
		&j.ErrorMessage, &j.StartedAt, &j.FetchedAt, &j.CompletedAt, &j.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get strava_sync_job: %w", err)
	}
	return &j, nil
}

// GetLinkedStravaActivitiesByUser loads all strava_activities for a user that are
// already linked to a local activity (i.e. activities.strava_activity_id points to them).
func (r *StravaRepository) GetLinkedStravaActivitiesByUser(ctx context.Context, userID uuid.UUID) ([]*model.StravaActivity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sa.id, sa.user_id, sa.strava_id, sa.title, sa.activity_type, sa.start_time, sa.distance,
		       sa.is_private, sa.is_flagged, sa.raw_json, sa.synced_at,
		       a.id AS linked_activity_id
		FROM strava_activities sa
		JOIN activities a ON a.strava_activity_id = sa.id
		WHERE sa.user_id = $1
		ORDER BY sa.start_time DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked strava_activities: %w", err)
	}
	defer rows.Close()

	var activities []*model.StravaActivity
	for rows.Next() {
		var a model.StravaActivity
		var rawJSON []byte
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.StravaID, &a.Title, &a.ActivityType, &a.StartTime, &a.Distance,
			&a.IsPrivate, &a.IsFlagged, &rawJSON, &a.SyncedAt,
			&a.LinkedActivityID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan strava_activity: %w", err)
		}
		if rawJSON != nil {
			json.Unmarshal(rawJSON, &a.RawJSON)
		}
		activities = append(activities, &a)
	}
	return activities, nil
}

// SetJobFetching transitions job to FETCHING.
func (r *StravaRepository) SetJobFetching(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE strava_sync_jobs SET status = $1 WHERE id = $2`,
		model.JobStatusFetching, id)
	return err
}

// SetJobDataFetched transitions job to DATA_FETCHED.
func (r *StravaRepository) SetJobDataFetched(ctx context.Context, id uuid.UUID, fetchedCount int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE strava_sync_jobs SET status = $1, fetched_count = $2, fetched_at = NOW() WHERE id = $3
	`, model.JobStatusDataFetched, fetchedCount, id)
	return err
}

// SetJobFetchFailed transitions job to FETCHING_FAILED.
func (r *StravaRepository) SetJobFetchFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE strava_sync_jobs SET status = $1, error_message = $2 WHERE id = $3
	`, model.JobStatusFetchingFailed, errMsg, id)
	return err
}

// SetJobProcessing transitions job to PROCESSING.
func (r *StravaRepository) SetJobProcessing(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE strava_sync_jobs SET status = $1 WHERE id = $2`,
		model.JobStatusProcessing, id)
	return err
}

// SetJobDataProcessed transitions job to DATA_PROCESSED.
func (r *StravaRepository) SetJobDataProcessed(ctx context.Context, id uuid.UUID, updatedCount, createdCount int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE strava_sync_jobs SET status = $1, updated_count = $2, created_count = $3, completed_at = NOW() WHERE id = $4
	`, model.JobStatusDataProcessed, updatedCount, createdCount, id)
	return err
}

// SetJobProcessFailed transitions job to PROCESSING_FAILED.
func (r *StravaRepository) SetJobProcessFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE strava_sync_jobs SET status = $1, error_message = $2 WHERE id = $3
	`, model.JobStatusProcessingFailed, errMsg, id)
	return err
}

// GetStravaActivitiesByJob retrieves all strava_activities tagged with the given job ID.
// It also populates LinkedActivityID from any local activity already linked to each row.
func (r *StravaRepository) GetStravaActivitiesByJob(ctx context.Context, jobID uuid.UUID) ([]*model.StravaActivity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sa.id, sa.user_id, sa.strava_id, sa.title, sa.activity_type, sa.start_time, sa.distance,
		       sa.is_private, sa.is_flagged, sa.raw_json, sa.synced_at,
		       a.id AS linked_activity_id
		FROM strava_activities sa
		LEFT JOIN activities a ON a.strava_activity_id = sa.id
		WHERE sa.job_id = $1
		ORDER BY sa.start_time DESC
	`, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query strava_activities by job: %w", err)
	}
	defer rows.Close()

	var activities []*model.StravaActivity
	for rows.Next() {
		var a model.StravaActivity
		var rawJSON []byte
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.StravaID, &a.Title, &a.ActivityType, &a.StartTime, &a.Distance,
			&a.IsPrivate, &a.IsFlagged, &rawJSON, &a.SyncedAt,
			&a.LinkedActivityID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan strava_activity: %w", err)
		}
		if rawJSON != nil {
			json.Unmarshal(rawJSON, &a.RawJSON)
		}
		activities = append(activities, &a)
	}
	return activities, nil
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
			strava_activity_id, created_at, updated_at
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
			&a.StravaActivityID, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, &a)
	}
	return activities, nil
}

// UpdateActivity updates name, sport, and strava_activity_id on a local activity.
func (r *ActivityRepository) UpdateActivity(ctx context.Context, id uuid.UUID, name, sport string, stravaActivityID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE activities
		SET name = $1, sport = $2, strava_activity_id = $3, updated_at = NOW()
		WHERE id = $4
	`, name, sport, stravaActivityID, id)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	return nil
}
