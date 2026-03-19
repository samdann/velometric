package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/velometric/backend/internal/model"
)

// ActivityPowerRecord is a single power data point with its activity ID, used for batch processing.
type ActivityPowerRecord struct {
	ActivityID uuid.UUID
	Timestamp  time.Time
	Power      int
}

// StatisticsRepository handles queries for the statistics endpoints.
type StatisticsRepository struct {
	pool *pgxpool.Pool
}

func NewStatisticsRepository(pool *pgxpool.Pool) *StatisticsRepository {
	return &StatisticsRepository{pool: pool}
}

// GetAvailablePowerYears returns distinct years (desc) where the user has activities with power data.
func (r *StatisticsRepository) GetAvailablePowerYears(ctx context.Context, userID uuid.UUID) ([]int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT EXTRACT(YEAR FROM start_time)::int AS year
		FROM activities
		WHERE user_id = $1 AND avg_power IS NOT NULL
		ORDER BY year DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available power years: %w", err)
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var year int
		if err := rows.Scan(&year); err != nil {
			return nil, fmt.Errorf("failed to scan year: %w", err)
		}
		years = append(years, year)
	}
	return years, nil
}

// GetAnnualMedianPowerCurve returns the median best power at each of the given durations for a year.
func (r *StatisticsRepository) GetAnnualMedianPowerCurve(ctx context.Context, userID uuid.UUID, year int, durations []int) ([]model.AnnualPowerCurvePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT apc.duration_seconds,
		       PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY apc.best_power)::int AS median_power
		FROM activity_power_curve apc
		JOIN activities a ON apc.activity_id = a.id
		WHERE a.user_id = $1
		  AND EXTRACT(YEAR FROM a.start_time) = $2
		  AND a.avg_power IS NOT NULL
		  AND apc.duration_seconds = ANY($3)
		GROUP BY apc.duration_seconds
		ORDER BY apc.duration_seconds
	`, userID, year, durations)
	if err != nil {
		return nil, fmt.Errorf("failed to get annual median power curve: %w", err)
	}
	defer rows.Close()

	var points []model.AnnualPowerCurvePoint
	for rows.Next() {
		var p model.AnnualPowerCurvePoint
		if err := rows.Scan(&p.DurationSeconds, &p.MedianPower); err != nil {
			return nil, fmt.Errorf("failed to scan power curve point: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}

// GetAnnualPowerRecords returns all power records for activities in a given year, ordered by (activity_id, timestamp).
func (r *StatisticsRepository) GetAnnualPowerRecords(ctx context.Context, userID uuid.UUID, year int) ([]ActivityPowerRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ar.activity_id, ar.timestamp, ar.power
		FROM activity_records ar
		JOIN activities a ON ar.activity_id = a.id
		WHERE a.user_id = $1
		  AND EXTRACT(YEAR FROM a.start_time) = $2
		  AND ar.power IS NOT NULL
		ORDER BY ar.activity_id, ar.timestamp
	`, userID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get annual power records: %w", err)
	}
	defer rows.Close()

	var records []ActivityPowerRecord
	for rows.Next() {
		var rec ActivityPowerRecord
		if err := rows.Scan(&rec.ActivityID, &rec.Timestamp, &rec.Power); err != nil {
			return nil, fmt.Errorf("failed to scan power record: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}
