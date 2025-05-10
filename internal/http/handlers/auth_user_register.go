package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/services"
	"github.com/tahsin005/affpilot-auth/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	if req.Email == "" || req.Password == "" || req.Username == "" || req.FirstName == "" || req.LastName == "" {
		utils.WriteError(w, http.StatusBadRequest, "Username, email, password, first name, and last name are required")
		return
	}

	// Check if email or username already exists
	var exists bool
	queryStatement := `
		SELECT EXISTS (SELECT 1 FROM users WHERE email = $1 OR username = $2)
	`
	err := database.DB.QueryRow(queryStatement, req.Email, req.Username).Scan(&exists)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		utils.WriteError(w, http.StatusConflict, "Email or username already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	userID := uuid.New()
	verificationToken := uuid.New().String()
	// tokenExpiry := time.Now().Add(5 * time.Minute)
	// now := time.Now()
	tokenExpiry := time.Now().UTC().Add(5 * time.Minute)
	now := time.Now().UTC()

	queryStatement = `
		INSERT INTO users (
			id, username, email, password_hash, first_name, last_name, email_verified,
			user_type, verification_token, token_expiry, deletion_requested, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, FALSE,
			'user', $7, $8, FALSE, FALSE, $9, $10
		)
	`
	_, err = database.DB.Exec(queryStatement, userID, req.Username, req.Email, string(hashedPassword), req.FirstName, req.LastName, verificationToken, tokenExpiry, now, now)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Assign role
	var roleID uuid.UUID
	queryStatement = `
		SELECT id FROM roles WHERE name = 'user'
	`
	err = database.DB.QueryRow(queryStatement).Scan(&roleID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch role")
		return
	}

	queryStatement = `
		INSERT INTO user_roles (user_id, role_id, assigned_by, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = database.DB.Exec(queryStatement, userID, roleID, userID, now)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to assign role")
		return
	}

	// Send verification email
	emailVerficationURL := os.Getenv("EMAIL_VERIFICATION_URL")
	if emailVerficationURL == "" {
		log.Println("EMAIL_VERIFICATION_URL environment variable not set, using default")
		emailVerficationURL = "http://localhost:8080/api/v1/auth/verify"
	}
	verificationLink := fmt.Sprintf("%s/%s", emailVerficationURL, verificationToken)
	emailBody := fmt.Sprintf(
		"Welcome to Affpilot!\n\nPlease verify your email by clicking the following link:\n%s\n\nThis link will expire in 2 hours.",
		verificationLink,
	)
	go services.SendEmail(req.Email, emailBody)

	userResp := models.RegisterUserResponse{
		ID:        userID,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UserType:  "user",
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully. Please check your email to verify your account.",
		"user":    userResp,
	})
}
