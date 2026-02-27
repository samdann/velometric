package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/velometric/backend/internal/model"
)

// ErrDuplicateActivity is returned when an activity with the same
// (user_id, start_time, sport, distance, duration) already exists.
var ErrDuplicateActivity = errors.New("activity already exists")

type ActivityRepository struct {
	pool *pgxpool.Pool
}

func NewActivityRepository(pool *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{pool: pool}
}

// Create inserts a new activity and returns its ID.
// Returns ErrDuplicateActivity if the same file was already uploaded.
func (r *ActivityRepository) Create(ctx context.Context, a *model.Activity) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO activities (
			user_id, name, sport, start_time, duration, distance, elevation_gain,
			avg_power, max_power, normalized_power, tss, intensity_factor, variability_index,
			avg_hr, max_hr, avg_cadence, max_cadence, avg_speed, max_speed,
			calories, avg_temperature, fit_file_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19,
			$20, $21, $22
		) ON CONFLICT ON CONSTRAINT unique_activity DO NOTHING
		RETURNING id
	`,
		a.UserID, a.Name, a.Sport, a.StartTime, a.Duration, a.Distance, a.ElevationGain,
		a.AvgPower, a.MaxPower, a.NormalizedPower, a.TSS, a.IntensityFactor, a.VariabilityIndex,
		a.AvgHR, a.MaxHR, a.AvgCadence, a.MaxCadence, a.AvgSpeed, a.MaxSpeed,
		a.Calories, a.AvgTemperature, a.FitFileURL,
	).Scan(&id)

	if err == pgx.ErrNoRows {
		return uuid.Nil, ErrDuplicateActivity
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create activity: %w", err)
	}
	return id, nil
}

// GetByID retrieves an activity by ID
func (r *ActivityRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Activity, error) {
	a := &model.Activity{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, sport, start_time, duration, distance, elevation_gain,
			avg_power, max_power, normalized_power, tss, intensity_factor, variability_index,
			avg_hr, max_hr, avg_cadence, max_cadence, avg_speed, max_speed,
			calories, avg_temperature, fit_file_url, created_at, updated_at
		FROM activities WHERE id = $1
	`, id).Scan(
		&a.ID, &a.UserID, &a.Name, &a.Sport, &a.StartTime, &a.Duration, &a.Distance, &a.ElevationGain,
		&a.AvgPower, &a.MaxPower, &a.NormalizedPower, &a.TSS, &a.IntensityFactor, &a.VariabilityIndex,
		&a.AvgHR, &a.MaxHR, &a.AvgCadence, &a.MaxCadence, &a.AvgSpeed, &a.MaxSpeed,
		&a.Calories, &a.AvgTemperature, &a.FitFileURL, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}
	return a, nil
}

// ListByUserID retrieves all activities for a user
func (r *ActivityRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Activity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, sport, start_time, duration, distance, elevation_gain,
			avg_power, max_power, normalized_power, tss, intensity_factor, variability_index,
			avg_hr, max_hr, avg_cadence, max_cadence, avg_speed, max_speed,
			calories, avg_temperature, fit_file_url, created_at, updated_at
		FROM activities
		WHERE user_id = $1
		ORDER BY start_time DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}
	defer rows.Close()

	var activities []*model.Activity
	for rows.Next() {
		a := &model.Activity{}
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Name, &a.Sport, &a.StartTime, &a.Duration, &a.Distance, &a.ElevationGain,
			&a.AvgPower, &a.MaxPower, &a.NormalizedPower, &a.TSS, &a.IntensityFactor, &a.VariabilityIndex,
			&a.AvgHR, &a.MaxHR, &a.AvgCadence, &a.MaxCadence, &a.AvgSpeed, &a.MaxSpeed,
			&a.Calories, &a.AvgTemperature, &a.FitFileURL, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, a)
	}
	return activities, nil
}

// InsertRecords bulk inserts activity records
func (r *ActivityRepository) InsertRecords(ctx context.Context, records []model.ActivityRecord) error {
	if len(records) == 0 {
		return nil
	}

	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"activity_records"},
		[]string{
			"activity_id", "timestamp", "lat", "lon", "altitude", "distance",
			"power", "heart_rate", "cadence", "speed", "temperature",
			"left_right_balance", "left_torque_effectiveness", "right_torque_effectiveness",
			"left_pedal_smoothness", "right_pedal_smoothness", "gradient",
		},
		pgx.CopyFromSlice(len(records), func(i int) ([]interface{}, error) {
			r := records[i]
			return []interface{}{
				r.ActivityID, r.Timestamp, r.Lat, r.Lon, r.Altitude, r.Distance,
				r.Power, r.HeartRate, r.Cadence, r.Speed, r.Temperature,
				r.LeftRightBalance, r.LeftTorqueEffectiveness, r.RightTorqueEffectiveness,
				r.LeftPedalSmoothness, r.RightPedalSmoothness, r.Gradient,
			}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to insert records: %w", err)
	}
	return nil
}

// InsertLaps bulk inserts activity laps
func (r *ActivityRepository) InsertLaps(ctx context.Context, laps []model.ActivityLap) error {
	if len(laps) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, lap := range laps {
		batch.Queue(`
			INSERT INTO activity_laps (
				activity_id, lap_number, start_time, duration, distance,
				avg_power, max_power, normalized_power, avg_hr, max_hr,
				avg_cadence, avg_speed, max_speed, ascent, descent, trigger
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		`,
			lap.ActivityID, lap.LapNumber, lap.StartTime, lap.Duration, lap.Distance,
			lap.AvgPower, lap.MaxPower, lap.NormalizedPower, lap.AvgHR, lap.MaxHR,
			lap.AvgCadence, lap.AvgSpeed, lap.MaxSpeed, lap.Ascent, lap.Descent, lap.Trigger,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range laps {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to insert lap: %w", err)
		}
	}
	return nil
}

// InsertPowerCurve bulk inserts power curve points
func (r *ActivityRepository) InsertPowerCurve(ctx context.Context, points []model.PowerCurvePoint) error {
	if len(points) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, p := range points {
		batch.Queue(`
			INSERT INTO activity_power_curve (activity_id, duration_seconds, best_power, avg_hr)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (activity_id, duration_seconds) DO UPDATE SET
				best_power = EXCLUDED.best_power,
				avg_hr = EXCLUDED.avg_hr
		`, p.ActivityID, p.DurationSeconds, p.BestPower, p.AvgHeartRate)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range points {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to insert power curve point: %w", err)
		}
	}
	return nil
}

// InsertEvents bulk inserts activity events
func (r *ActivityRepository) InsertEvents(ctx context.Context, events []model.ActivityEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, e := range events {
		dataJSON, _ := json.Marshal(e.Data)
		batch.Queue(`
			INSERT INTO activity_events (activity_id, timestamp, event_type, data)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (activity_id, timestamp, event_type) DO NOTHING
		`, e.ActivityID, e.Timestamp, e.EventType, dataJSON)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range events {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}
	return nil
}

// GetRecords retrieves all records for an activity
func (r *ActivityRepository) GetRecords(ctx context.Context, activityID uuid.UUID) ([]model.ActivityRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT activity_id, timestamp, lat, lon, altitude, distance,
			power, heart_rate, cadence, speed, temperature,
			left_right_balance, left_torque_effectiveness, right_torque_effectiveness,
			left_pedal_smoothness, right_pedal_smoothness, gradient
		FROM activity_records
		WHERE activity_id = $1
		ORDER BY timestamp ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []model.ActivityRecord
	for rows.Next() {
		var r model.ActivityRecord
		err := rows.Scan(
			&r.ActivityID, &r.Timestamp, &r.Lat, &r.Lon, &r.Altitude, &r.Distance,
			&r.Power, &r.HeartRate, &r.Cadence, &r.Speed, &r.Temperature,
			&r.LeftRightBalance, &r.LeftTorqueEffectiveness, &r.RightTorqueEffectiveness,
			&r.LeftPedalSmoothness, &r.RightPedalSmoothness, &r.Gradient,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		records = append(records, r)
	}
	return records, nil
}

// GetLaps retrieves all laps for an activity
func (r *ActivityRepository) GetLaps(ctx context.Context, activityID uuid.UUID) ([]model.ActivityLap, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, activity_id, lap_number, start_time, duration, distance,
			avg_power, max_power, normalized_power, avg_hr, max_hr,
			avg_cadence, avg_speed, max_speed, ascent, descent, trigger
		FROM activity_laps
		WHERE activity_id = $1
		ORDER BY lap_number ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get laps: %w", err)
	}
	defer rows.Close()

	var laps []model.ActivityLap
	for rows.Next() {
		var l model.ActivityLap
		if err := rows.Scan(
			&l.ID, &l.ActivityID, &l.LapNumber, &l.StartTime, &l.Duration, &l.Distance,
			&l.AvgPower, &l.MaxPower, &l.NormalizedPower, &l.AvgHR, &l.MaxHR,
			&l.AvgCadence, &l.AvgSpeed, &l.MaxSpeed, &l.Ascent, &l.Descent, &l.Trigger,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lap: %w", err)
		}
		laps = append(laps, l)
	}
	return laps, nil
}

// GetPowerCurve retrieves the power curve for an activity
func (r *ActivityRepository) GetPowerCurve(ctx context.Context, activityID uuid.UUID) ([]model.PowerCurvePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT activity_id, duration_seconds, best_power, avg_hr
		FROM activity_power_curve
		WHERE activity_id = $1
		ORDER BY duration_seconds ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get power curve: %w", err)
	}
	defer rows.Close()

	var points []model.PowerCurvePoint
	for rows.Next() {
		var p model.PowerCurvePoint
		if err := rows.Scan(&p.ActivityID, &p.DurationSeconds, &p.BestPower, &p.AvgHeartRate); err != nil {
			return nil, fmt.Errorf("failed to scan power curve point: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}
