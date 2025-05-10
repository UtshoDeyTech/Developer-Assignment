package database

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateSystemAdminIfNotExists() {
	username := strings.TrimSpace(os.Getenv("SYSTEM_ADMIN_USERNAME"))
	password := os.Getenv("SYSTEM_ADMIN_PASSWORD")
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SYSTEM_ADMIN_EMAIL")))

	if username == "" || password == "" || email == "" {
		log.Println("System admin environment variables are missing. Skipping admin creation.")
		return
	}

	var exists bool
	queryStatement := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1 OR username = $2)`
	err := DB.QueryRow(queryStatement, email, username).Scan(&exists)
	if err != nil {
		log.Println("Error checking for existing admin:", err)
		return
	}

	if exists {
		log.Println("System admin already exists. Skipping creation.")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing system admin password:", err)
		return
	}

	adminID := uuid.New()
	now := time.Now()
	verificationToken := uuid.NewString()
	tokenExpiry := now.AddDate(5, 0, 0)

	queryStatement = `
		INSERT INTO users (
			id, username, email, password_hash, first_name, last_name, email_verified,
			user_type, deletion_requested, active, created_at, updated_at,
			verification_token, token_expiry
		) VALUES (
			$1, $2, $3, $4, 'Admin', 'User', TRUE,
			'system_admin', FALSE, TRUE, $5, $5,
			$6, $7
		)
	`
	_, err = DB.Exec(queryStatement, adminID, username, email, string(hashedPassword), now, verificationToken, tokenExpiry)
	if err != nil {
		log.Println("Failed to create system admin:", err)
		return
	}

	var roleID uuid.UUID
	queryStatement = `SELECT id FROM roles WHERE name = 'system_admin'`
	err = DB.QueryRow(queryStatement).Scan(&roleID)
	if err != nil {
		log.Println("Failed to fetch admin role:", err)
		return
	}

	queryStatement = `INSERT INTO user_roles (user_id, role_id, assigned_by, created_at) VALUES ($1, $2, $3, $4)`
	_, err = DB.Exec(queryStatement, adminID, roleID, adminID, now)
	if err != nil {
		log.Println("Failed to assign system admin role:", err)
		return
	}

	log.Println("System admin created successfully.")
}
