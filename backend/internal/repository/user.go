package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/velometric/backend/internal/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetFirst(ctx context.Context) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, name, ftp, max_hr, weight, created_at, updated_at
		FROM users
		ORDER BY created_at ASC
		LIMIT 1
	`).Scan(&u.ID, &u.Email, &u.Name, &u.FTP, &u.MaxHR, &u.Weight, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name, email string, weight *float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET name = $1, email = $2, weight = $3, updated_at = NOW()
		WHERE id = $4
	`, name, email, weight, id)
	return err
}

func (r *UserRepository) UpdateMaxHR(ctx context.Context, id uuid.UUID, maxHR int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET max_hr = $1, updated_at = NOW() WHERE id = $2
	`, maxHR, id)
	return err
}

func (r *UserRepository) UpdateFTP(ctx context.Context, id uuid.UUID, ftp int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET ftp = $1, updated_at = NOW() WHERE id = $2
	`, ftp, id)
	return err
}

func (r *UserRepository) GetHRZones(ctx context.Context, userID uuid.UUID) ([]model.HRZone, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT zone_number, name, min_percentage, max_percentage, color
		FROM user_hr_zones
		WHERE user_id = $1
		ORDER BY zone_number ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	zones := make([]model.HRZone, 0)
	for rows.Next() {
		var z model.HRZone
		var color *string
		if err := rows.Scan(&z.ZoneNumber, &z.Name, &z.MinPercentage, &z.MaxPercentage, &color); err != nil {
			return nil, err
		}
		if color != nil {
			z.Color = *color
		}
		zones = append(zones, z)
	}
	return zones, rows.Err()
}

func (r *UserRepository) UpsertHRZones(ctx context.Context, userID uuid.UUID, zones []model.HRZone) error {
	batch := &pgx.Batch{}
	for _, z := range zones {
		var colorPtr *string
		if z.Color != "" {
			c := z.Color
			colorPtr = &c
		}
		batch.Queue(`
			INSERT INTO user_hr_zones (user_id, zone_number, name, min_percentage, max_percentage, color)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, zone_number) DO UPDATE
			SET name = EXCLUDED.name,
				min_percentage = EXCLUDED.min_percentage,
				max_percentage = EXCLUDED.max_percentage,
				color = EXCLUDED.color
		`, userID, z.ZoneNumber, z.Name, z.MinPercentage, z.MaxPercentage, colorPtr)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range zones {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepository) GetPowerZones(ctx context.Context, userID uuid.UUID) ([]model.PowerZone, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT zone_number, name, min_percentage, max_percentage, color
		FROM user_power_zones
		WHERE user_id = $1
		ORDER BY zone_number ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	zones := make([]model.PowerZone, 0)
	for rows.Next() {
		var z model.PowerZone
		var color *string
		if err := rows.Scan(&z.ZoneNumber, &z.Name, &z.MinPercentage, &z.MaxPercentage, &color); err != nil {
			return nil, err
		}
		if color != nil {
			z.Color = *color
		}
		zones = append(zones, z)
	}
	return zones, rows.Err()
}

func (r *UserRepository) UpsertPowerZones(ctx context.Context, userID uuid.UUID, zones []model.PowerZone) error {
	batch := &pgx.Batch{}
	for _, z := range zones {
		var colorPtr *string
		if z.Color != "" {
			c := z.Color
			colorPtr = &c
		}
		batch.Queue(`
			INSERT INTO user_power_zones (user_id, zone_number, name, min_percentage, max_percentage, color)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, zone_number) DO UPDATE
			SET name = EXCLUDED.name,
				min_percentage = EXCLUDED.min_percentage,
				max_percentage = EXCLUDED.max_percentage,
				color = EXCLUDED.color
		`, userID, z.ZoneNumber, z.Name, z.MinPercentage, z.MaxPercentage, colorPtr)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range zones {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}
