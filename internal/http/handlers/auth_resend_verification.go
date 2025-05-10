package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/services"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.ResendVerificationEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if !utils.IsValidEmail(req.Email) {
		utils.WriteError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	var userID, verificationToken string
	var active, emailVerified bool
	query := `
		SELECT id, verification_token, active, email_verified
		FROM users
		WHERE email = $1
	`
	err := database.DB.QueryRow(query, req.Email).Scan(&userID, &verificationToken, &active, &emailVerified)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If the email exists, a verification link has been sent",
		})
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if emailVerified {
		utils.WriteError(w, http.StatusBadRequest, "Email is already verified")
		return
	}

	verificationToken = uuid.New().String()

	tokenExpiry := time.Now().UTC().Add(5 * time.Minute)
	updateQuery := `
		UPDATE users
		SET verification_token = $1, token_expiry = $2, updated_at = $3
		WHERE id = $4
	`
	_, err = database.DB.Exec(updateQuery, verificationToken, tokenExpiry, time.Now(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update verification token")
		return
	}

	// Send verification email
	emailVerificationURL := os.Getenv("EMAIL_VERIFICATION_URL")
	if emailVerificationURL == "" {
		log.Println("EMAIL_VERIFICATION_URL environment variable not set, using default")
		emailVerificationURL = "http://localhost:8080/api/v1/auth/verify"
	}
	verificationLink := fmt.Sprintf("%s/%s", emailVerificationURL, verificationToken)
	emailBody := fmt.Sprintf(
		"Welcome to Affpilot!\n\nPlease verify your email by clicking the following link:\n%s\n\nThis link will expire in 5 minutes.",
		verificationLink,
	)
	go services.SendEmail(req.Email, emailBody)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If the email exists, a verification link has been sent",
	})
}