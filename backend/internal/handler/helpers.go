package handler

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/velometric/backend/internal/database"
)

// GetDemoUserID is the exported version of getDemoUserID for use outside the package.
func GetDemoUserID(ctx context.Context, db *database.DB) (uuid.UUID, error) {
	return getDemoUserID(ctx, db)
}

// getDemoUserID returns the demo user ID for development
// TODO: Replace with proper auth
func getDemoUserID(ctx context.Context, db *database.DB) (uuid.UUID, error) {
	var userID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		SELECT id FROM users WHERE email = $1
	`, "demo@velometric.app").Scan(&userID)

	if err != nil {
		// Create demo user if it doesn't exist
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO users (email, name, ftp, max_hr, weight)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
			RETURNING id
		`, "demo@velometric.app", "Demo Cyclist", 250, 185, 75.0).Scan(&userID)

		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get or create demo user: %w", err)
		}
	}

	return userID, nil
}
