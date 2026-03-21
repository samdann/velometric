package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/velometric/backend/internal/model"
)

// sanitizeFloat sets NaN/Inf float pointers to nil so they serialize to JSON null.
func sanitizeFloat(f *float64) *float64 {
	if f == nil || math.IsNaN(*f) || math.IsInf(*f, 0) {
		return nil
	}
	return f
}

// clampGradient discards physically impossible gradient spikes (corrupt FIT data).
// Real-world gradients stay well within ±100%; anything beyond ±200% is noise.
func clampGradient(f *float64) *float64 {
	if f == nil || math.IsNaN(*f) || math.IsInf(*f, 0) {
		return nil
	}
	if *f > 200 || *f < -200 {
		return nil
	}
	return f
}

// ErrDuplicateActivity is returned when an activity with the same
// (user_id, start_time, sport, distance, duration) already exists.
var ErrDuplicateActivity = errors.New("activity already exists")
var ErrActivityNotFound = errors.New("activity not found")

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
			calories, avg_temperature, fit_file_url, device_name, location
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24
		) ON CONFLICT ON CONSTRAINT unique_activity DO NOTHING
		RETURNING id
	`,
		a.UserID, a.Name, a.Sport, a.StartTime, a.Duration, a.Distance, a.ElevationGain,
		a.AvgPower, a.MaxPower, a.NormalizedPower, a.TSS, a.IntensityFactor, a.VariabilityIndex,
		a.AvgHR, a.MaxHR, a.AvgCadence, a.MaxCadence, a.AvgSpeed, a.MaxSpeed,
		a.Calories, a.AvgTemperature, a.FitFileURL, a.DeviceName, a.Location,
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
			calories, avg_temperature, fit_file_url, device_name, location,
			strava_activity_id, created_at, updated_at
		FROM activities WHERE id = $1
	`, id).Scan(
		&a.ID, &a.UserID, &a.Name, &a.Sport, &a.StartTime, &a.Duration, &a.Distance, &a.ElevationGain,
		&a.AvgPower, &a.MaxPower, &a.NormalizedPower, &a.TSS, &a.IntensityFactor, &a.VariabilityIndex,
		&a.AvgHR, &a.MaxHR, &a.AvgCadence, &a.MaxCadence, &a.AvgSpeed, &a.MaxSpeed,
		&a.Calories, &a.AvgTemperature, &a.FitFileURL, &a.DeviceName, &a.Location,
		&a.StravaActivityID, &a.CreatedAt, &a.UpdatedAt,
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
			calories, avg_temperature, fit_file_url, device_name, location,
			strava_activity_id, created_at, updated_at
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
			&a.Calories, &a.AvgTemperature, &a.FitFileURL, &a.DeviceName, &a.Location,
			&a.StravaActivityID, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, a)
	}
	return activities, nil
}

// sortColumnMap maps allowed sort keys to SQL column names.
var sortColumnMap = map[string]string{
	"date":      "start_time",
	"distance":  "distance",
	"duration":  "duration",
	"elevation": "elevation_gain",
}

func (r *ActivityRepository) ListByUserIDPaginated(ctx context.Context, userID uuid.UUID, page, limit int, f model.ActivityFilter) ([]*model.Activity, int, error) {
	offset := (page - 1) * limit

	// Safe ORDER BY — allowlist only
	col, ok := sortColumnMap[f.SortBy]
	if !ok {
		col = "start_time"
	}
	dir := "DESC"
	if f.SortOrder == "asc" {
		dir = "ASC"
	}

	// Distance filter is stored in metres in DB; inputs are km
	var distMin, distMax float64
	distMin = 0
	distMax = 1e9
	if f.DistanceMinKm != nil {
		distMin = *f.DistanceMinKm * 1000
	}
	if f.DistanceMaxKm != nil {
		distMax = *f.DistanceMaxKm * 1000
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, name, sport, start_time, duration, distance, elevation_gain,
			avg_power, max_power, normalized_power, tss, intensity_factor, variability_index,
			avg_hr, max_hr, avg_cadence, max_cadence, avg_speed, max_speed,
			calories, avg_temperature, fit_file_url, device_name, location,
			strava_activity_id, created_at, updated_at,
			COUNT(*) OVER () AS total_count
		FROM activities
		WHERE user_id = $1
			AND ($2 = '' OR name ILIKE '%%' || $2 || '%%')
			AND ($3 = '' OR LOWER(sport) = LOWER($3))
			AND start_time >= $4
			AND start_time <= $5
			AND distance >= $6
			AND distance <= $7
		ORDER BY %s %s
		LIMIT $8 OFFSET $9
	`, col, dir)

	rows, err := r.pool.Query(ctx, query,
		userID, f.Query, f.Sport,
		f.DateFrom, f.DateTo,
		distMin, distMax,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list activities: %w", err)
	}
	defer rows.Close()

	var activities []*model.Activity
	var total int
	for rows.Next() {
		a := &model.Activity{}
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Name, &a.Sport, &a.StartTime, &a.Duration, &a.Distance, &a.ElevationGain,
			&a.AvgPower, &a.MaxPower, &a.NormalizedPower, &a.TSS, &a.IntensityFactor, &a.VariabilityIndex,
			&a.AvgHR, &a.MaxHR, &a.AvgCadence, &a.MaxCadence, &a.AvgSpeed, &a.MaxSpeed,
			&a.Calories, &a.AvgTemperature, &a.FitFileURL, &a.DeviceName, &a.Location,
			&a.StravaActivityID, &a.CreatedAt, &a.UpdatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, a)
	}
	if activities == nil {
		activities = make([]*model.Activity, 0)
	}
	return activities, total, nil
}

// GetDistinctSports returns the distinct sport values for a user, sorted alphabetically.
func (r *ActivityRepository) GetDistinctSports(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT sport FROM activities WHERE user_id = $1 ORDER BY sport
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query distinct sports: %w", err)
	}
	defer rows.Close()

	sports := make([]string, 0)
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("failed to scan sport: %w", err)
		}
		sports = append(sports, s)
	}
	return sports, nil
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
				r.LeftPedalSmoothness, r.RightPedalSmoothness, clampGradient(r.Gradient),
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
			INSERT INTO activity_power_curve (activity_id, duration_seconds, best_power, avg_hr, avg_speed, avg_gradient, avg_cadence, avg_lr_balance, avg_torque_effectiveness)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (activity_id, duration_seconds) DO UPDATE SET
				best_power = EXCLUDED.best_power,
				avg_hr = EXCLUDED.avg_hr,
				avg_speed = EXCLUDED.avg_speed,
				avg_gradient = EXCLUDED.avg_gradient,
				avg_cadence = EXCLUDED.avg_cadence,
				avg_lr_balance = EXCLUDED.avg_lr_balance,
				avg_torque_effectiveness = EXCLUDED.avg_torque_effectiveness
		`, p.ActivityID, p.DurationSeconds, p.BestPower, p.AvgHeartRate,
			p.AvgSpeed, p.AvgGradient, p.AvgCadence, p.AvgLRBalance, p.AvgTorqueEffectiveness)
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
		r.Lat = sanitizeFloat(r.Lat)
		r.Lon = sanitizeFloat(r.Lon)
		r.Altitude = sanitizeFloat(r.Altitude)
		r.Distance = sanitizeFloat(r.Distance)
		r.Speed = sanitizeFloat(r.Speed)
		r.Temperature = sanitizeFloat(r.Temperature)
		r.LeftRightBalance = sanitizeFloat(r.LeftRightBalance)
		r.LeftTorqueEffectiveness = sanitizeFloat(r.LeftTorqueEffectiveness)
		r.RightTorqueEffectiveness = sanitizeFloat(r.RightTorqueEffectiveness)
		r.LeftPedalSmoothness = sanitizeFloat(r.LeftPedalSmoothness)
		r.RightPedalSmoothness = sanitizeFloat(r.RightPedalSmoothness)
		r.Gradient = sanitizeFloat(r.Gradient)
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

	laps := make([]model.ActivityLap, 0)
	for rows.Next() {
		var l model.ActivityLap
		if err := rows.Scan(
			&l.ID, &l.ActivityID, &l.LapNumber, &l.StartTime, &l.Duration, &l.Distance,
			&l.AvgPower, &l.MaxPower, &l.NormalizedPower, &l.AvgHR, &l.MaxHR,
			&l.AvgCadence, &l.AvgSpeed, &l.MaxSpeed, &l.Ascent, &l.Descent, &l.Trigger,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lap: %w", err)
		}
		l.AvgSpeed = sanitizeFloat(l.AvgSpeed)
		l.MaxSpeed = sanitizeFloat(l.MaxSpeed)
		l.Ascent = sanitizeFloat(l.Ascent)
		l.Descent = sanitizeFloat(l.Descent)
		laps = append(laps, l)
	}
	return laps, nil
}

// GetElevationProfile retrieves distance/altitude/temperature points for an activity
func (r *ActivityRepository) GetElevationProfile(ctx context.Context, activityID uuid.UUID) ([]model.ElevationPoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT distance, altitude, temperature
		FROM activity_records
		WHERE activity_id = $1
		  AND distance IS NOT NULL
		  AND altitude IS NOT NULL
		ORDER BY timestamp ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elevation profile: %w", err)
	}
	defer rows.Close()

	points := make([]model.ElevationPoint, 0)
	for rows.Next() {
		var p model.ElevationPoint
		var distMeters float64
		if err := rows.Scan(&distMeters, &p.Altitude, &p.Temperature); err != nil {
			return nil, fmt.Errorf("failed to scan elevation point: %w", err)
		}
		if sanitizeFloat(&p.Altitude) == nil {
			continue
		}
		p.Temperature = sanitizeFloat(p.Temperature)
		p.Distance = distMeters / 1000.0 // convert to km
		points = append(points, p)
	}
	return points, nil
}

// GetSpeedProfile retrieves raw distance/speed/power points for an activity
func (r *ActivityRepository) GetSpeedProfile(ctx context.Context, activityID uuid.UUID) ([]model.SpeedPoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT distance, speed, power
		FROM activity_records
		WHERE activity_id = $1
		  AND distance IS NOT NULL
		  AND speed IS NOT NULL
		ORDER BY timestamp ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get speed profile: %w", err)
	}
	defer rows.Close()

	points := make([]model.SpeedPoint, 0)
	for rows.Next() {
		var distMeters, speedMps float64
		var power *float64
		if err := rows.Scan(&distMeters, &speedMps, &power); err != nil {
			return nil, fmt.Errorf("failed to scan speed point: %w", err)
		}
		if sanitizeFloat(&speedMps) == nil {
			continue
		}
		p := model.SpeedPoint{
			Distance: distMeters / 1000.0,
			Speed:    speedMps * 3.6,
			Power:    sanitizeFloat(power),
		}
		points = append(points, p)
	}
	return points, nil
}

// GetHRCadenceProfile retrieves raw distance/heart_rate/cadence points for an activity
func (r *ActivityRepository) GetHRCadenceProfile(ctx context.Context, activityID uuid.UUID) ([]model.HRCadencePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT distance, heart_rate, cadence
		FROM activity_records
		WHERE activity_id = $1
		  AND distance IS NOT NULL
		  AND (heart_rate IS NOT NULL OR cadence IS NOT NULL)
		ORDER BY timestamp ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hr/cadence profile: %w", err)
	}
	defer rows.Close()

	points := make([]model.HRCadencePoint, 0)
	for rows.Next() {
		var distMeters float64
		var hr, cadence *int
		if err := rows.Scan(&distMeters, &hr, &cadence); err != nil {
			return nil, fmt.Errorf("failed to scan hr/cadence point: %w", err)
		}
		points = append(points, model.HRCadencePoint{
			Distance:  distMeters / 1000.0,
			HeartRate: hr,
			Cadence:   cadence,
		})
	}
	return points, nil
}

// GetRoute retrieves GPS coordinates for an activity
func (r *ActivityRepository) GetRoute(ctx context.Context, activityID uuid.UUID) ([]model.RoutePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT lat, lon, distance
		FROM activity_records
		WHERE activity_id = $1
		  AND lat IS NOT NULL
		  AND lon IS NOT NULL
		ORDER BY timestamp ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	defer rows.Close()

	points := make([]model.RoutePoint, 0)
	for rows.Next() {
		var p model.RoutePoint
		var distMeters *float64
		if err := rows.Scan(&p.Lat, &p.Lon, &distMeters); err != nil {
			return nil, fmt.Errorf("failed to scan route point: %w", err)
		}
		if distMeters != nil {
			km := *distMeters / 1000.0
			p.Distance = &km
		}
		points = append(points, p)
	}
	return points, nil
}

// GetPowerCurve retrieves the power curve for an activity
func (r *ActivityRepository) GetPowerCurve(ctx context.Context, activityID uuid.UUID) ([]model.PowerCurvePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT activity_id, duration_seconds, best_power, avg_hr,
		       avg_speed, avg_gradient, avg_cadence, avg_lr_balance, avg_torque_effectiveness
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
		if err := rows.Scan(
			&p.ActivityID, &p.DurationSeconds, &p.BestPower, &p.AvgHeartRate,
			&p.AvgSpeed, &p.AvgGradient, &p.AvgCadence, &p.AvgLRBalance, &p.AvgTorqueEffectiveness,
		); err != nil {
			return nil, fmt.Errorf("failed to scan power curve point: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}

// HRTimePoint is a lightweight timestamp + heart_rate pair for zone computation
type HRTimePoint struct {
	Timestamp time.Time
	HeartRate int
}

// GetHRTimeSeries retrieves ordered timestamp+heart_rate pairs for an activity
func (r *ActivityRepository) GetHRTimeSeries(ctx context.Context, activityID uuid.UUID) ([]HRTimePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT timestamp, heart_rate
		FROM activity_records
		WHERE activity_id = $1 AND heart_rate IS NOT NULL
		ORDER BY timestamp ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get HR time series: %w", err)
	}
	defer rows.Close()

	var points []HRTimePoint
	for rows.Next() {
		var p HRTimePoint
		if err := rows.Scan(&p.Timestamp, &p.HeartRate); err != nil {
			return nil, fmt.Errorf("failed to scan HR time point: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}

// PowerTimePoint is a lightweight timestamp + power pair for zone computation
type PowerTimePoint struct {
	Timestamp time.Time
	Power     int
}

// GetPowerTimeSeries retrieves ordered timestamp+power pairs for an activity
func (r *ActivityRepository) GetPowerTimeSeries(ctx context.Context, activityID uuid.UUID) ([]PowerTimePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT timestamp, power
		FROM activity_records
		WHERE activity_id = $1 AND power IS NOT NULL
		ORDER BY timestamp ASC
	`, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get power time series: %w", err)
	}
	defer rows.Close()

	var points []PowerTimePoint
	for rows.Next() {
		var p PowerTimePoint
		if err := rows.Scan(&p.Timestamp, &p.Power); err != nil {
			return nil, fmt.Errorf("failed to scan power time point: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}

// GetFeedActivities returns paginated activities with embedded mini-routes for the dashboard feed.
// Routes are downsampled to at most 50 points per activity.
func (r *ActivityRepository) GetFeedActivities(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.FeedActivity, int, error) {
	offset := (page - 1) * limit

	// Step 1: fetch paginated activity rows joined with user name.
	type actRow struct {
		id             uuid.UUID
		userName       string
		startTime      time.Time
		deviceName     *string
		location       *string
		name           string
		distanceMeters float64
		duration       int
		elevationGain  float64
		total          int
	}

	rows, err := r.pool.Query(ctx, `
		SELECT a.id, u.name, a.start_time, a.device_name, a.location,
		       a.name, a.distance, a.duration, a.elevation_gain,
		       COUNT(*) OVER () AS total_count
		FROM activities a
		JOIN users u ON u.id = a.user_id
		WHERE a.user_id = $1
		ORDER BY a.start_time DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query feed activities: %w", err)
	}
	defer rows.Close()

	var actRows []actRow
	var total int
	for rows.Next() {
		var ar actRow
		if err := rows.Scan(
			&ar.id, &ar.userName, &ar.startTime, &ar.deviceName, &ar.location,
			&ar.name, &ar.distanceMeters, &ar.duration, &ar.elevationGain,
			&ar.total,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan feed activity: %w", err)
		}
		total = ar.total
		actRows = append(actRows, ar)
	}
	rows.Close()

	if len(actRows) == 0 {
		return make([]model.FeedActivity, 0), total, nil
	}

	// Step 2: collect IDs and batch-fetch mini-routes (sampled to ≤50 pts each).
	ids := make([]uuid.UUID, len(actRows))
	for i, ar := range actRows {
		ids[i] = ar.id
	}

	routeRows, err := r.pool.Query(ctx, `
		SELECT activity_id, lat, lon
		FROM (
			SELECT activity_id, lat, lon,
			       ROW_NUMBER() OVER (PARTITION BY activity_id ORDER BY timestamp ASC) AS rn,
			       COUNT(*) OVER (PARTITION BY activity_id) AS total
			FROM activity_records
			WHERE activity_id = ANY($1)
			  AND lat IS NOT NULL
			  AND lon IS NOT NULL
		) sub
		WHERE MOD(rn - 1, GREATEST(total / 50, 1)) = 0
		ORDER BY activity_id, rn
	`, ids)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query mini-routes: %w", err)
	}
	defer routeRows.Close()

	routeMap := make(map[uuid.UUID][]model.RoutePoint)
	for routeRows.Next() {
		var aid uuid.UUID
		var p model.RoutePoint
		if err := routeRows.Scan(&aid, &p.Lat, &p.Lon); err != nil {
			return nil, 0, fmt.Errorf("failed to scan route point: %w", err)
		}
		routeMap[aid] = append(routeMap[aid], p)
	}

	// Step 3: assemble.
	feed := make([]model.FeedActivity, len(actRows))
	for i, ar := range actRows {
		pts := routeMap[ar.id]
		if pts == nil {
			pts = make([]model.RoutePoint, 0)
		}
		feed[i] = model.FeedActivity{
			ID:              ar.id,
			UserName:        ar.userName,
			StartTime:       ar.startTime,
			DeviceName:      ar.deviceName,
			Location:        ar.location,
			Name:            ar.name,
			DistanceKm:      ar.distanceMeters / 1000.0,
			DurationSeconds: ar.duration,
			ElevationGainM:  ar.elevationGain,
			Route:           pts,
		}
	}
	return feed, total, nil
}

// GetFirstGPSPoint returns the first lat/lon from activity_records for lazy geocoding.
func (r *ActivityRepository) GetFirstGPSPoint(ctx context.Context, activityID uuid.UUID) (float64, float64, error) {
	var lat, lon float64
	err := r.pool.QueryRow(ctx, `
		SELECT lat, lon FROM activity_records
		WHERE activity_id = $1 AND lat IS NOT NULL AND lon IS NOT NULL
		ORDER BY timestamp ASC LIMIT 1
	`, activityID).Scan(&lat, &lon)
	if err != nil {
		return 0, 0, fmt.Errorf("no GPS data: %w", err)
	}
	return lat, lon, nil
}

// UpdateActivityLocation persists a resolved location string.
func (r *ActivityRepository) UpdateActivityLocation(ctx context.Context, activityID uuid.UUID, location string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE activities SET location = $1 WHERE id = $2`,
		location, activityID,
	)
	return err
}

func (r *ActivityRepository) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM activities WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrActivityNotFound
	}
	return nil
}
