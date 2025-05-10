package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func UsersListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queryStatement := `
		SELECT id, username, email, first_name, last_name, email_verified, user_type, 
		verification_token, token_expiry, deletion_requested, active, created_at, updated_at
		FROM users
	`

	rows, err := database.DB.Query(queryStatement)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Database query error")
		log.Println("Query error:", err)
		return
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		var tokenExpiry sql.NullTime
		err := rows.Scan(
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
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to scan user")
			log.Println("Scan error:", err)
			return
		}

		if tokenExpiry.Valid {
			user.TokenExpiry = &tokenExpiry.Time
		}

		users = append(users, user)
	}


	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User's List",
		"users": users,
	})
}