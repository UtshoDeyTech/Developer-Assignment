package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func UserDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	userIDStr := vars["user_id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user_id format")
		return
	}

	query := `
		SELECT id, username, email, first_name, last_name, email_verified, user_type, 
		verification_token, token_expiry, deletion_requested, active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	var tokenExpiry sql.NullTime

	err = database.DB.QueryRow(query, userID).Scan(
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
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, "User not found")
		} else {
			utils.WriteError(w, http.StatusInternalServerError, "Database query error")
			log.Println("Query error:", err)
		}
		return
	}

	if tokenExpiry.Valid {
		user.TokenExpiry = &tokenExpiry.Time
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User details fetched successfully",
		"user":    user,
	})
}