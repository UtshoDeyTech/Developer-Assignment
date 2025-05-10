package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	token := vars["verification_token"]
	if token == "" {
		utils.WriteError(w, http.StatusBadRequest, "Verification token is required")
		return
	}

	log.Printf("Verfiying token: %s", token)

	var userID string
	var tokenExpiry time.Time
	queryStatement := `
		SELECT id, token_expiry
		FROM users
		WHERE verification_token = $1 AND email_verified = FALSE
	`

	err := database.DB.QueryRow(queryStatement, token).Scan(&userID, &tokenExpiry)
	if err == sql.ErrNoRows {
		utils.WriteError(w, http.StatusBadRequest, "Invalid or already used verification token")
		return
	}
	if err != nil {
		log.Printf("Database error verifying token: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "Database error")
		return
	}


	if time.Now().UTC().After(tokenExpiry) {
		utils.WriteError(w, http.StatusBadRequest, "Verification token has expired")
		return
	} else {
		log.Printf("Token expiry: %v", tokenExpiry)
		log.Println(time.Now())
	}

	// Update user to verified and active
	queryStatement = `
		UPDATE users
		SET email_verified = TRUE, active = TRUE, verification_token = 'verified', token_expiry = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err = database.DB.Exec(queryStatement, userID)
	if err != nil {
		log.Printf("Failed to update user verification: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	loginURL := os.Getenv("LOGIN_URL")
	if loginURL == "" {
		log.Println("LOGIN_URL environment variable not set, using default")
		loginURL = "https://github.com/tahsin005"
	}

	http.Redirect(w, r, loginURL, http.StatusFound)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verification successful",
	})
}