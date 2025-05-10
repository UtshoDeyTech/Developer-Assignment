package utils

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/models"
)

func GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	var tokenExpiry sql.NullTime
	query := `
		SELECT id, username, email, first_name, last_name, email_verified, user_type, verification_token, token_expiry, deletion_requested, active, created_at, updated_at
		FROM users
		WHERE id = $1 AND active = TRUE
	`
	err := database.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.EmailVerified,
		&user.UserType,
		&user.VerificationToken,
		&tokenExpiry,
		&user.DeletionRequested,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	if tokenExpiry.Valid {
		user.TokenExpiry = &tokenExpiry.Time
	}

	return &user, nil
}