package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func UserRoleDemoteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	userIDStr := vars["user_id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user_id format")
		return
	}

	if userIDStr == claims.UserID {
		utils.WriteError(w, http.StatusForbidden, "Forbidden: You cannot demote yourself")
		return
	}

	var req models.RoleDemoteRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	validRoles := map[string]bool {
		"moderator": true,
		"user": true,
	}

	if !validRoles[req.Role] {
		utils.WriteError(w, http.StatusBadRequest, "Invalid role: only accepts moderator or user")
		return
	}

	user, err := utils.GetUserByID(userID)
	if err == sql.ErrNoRows {
		utils.WriteError(w, http.StatusNotFound, "User not found or already deleted")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if user.UserType == "system_admin" {
		utils.WriteError(w, http.StatusForbidden, "Forbidden: System admins cannot be demoted")
		return
	}

	if user.UserType == "user" {
		utils.WriteError(w, http.StatusBadRequest, "User is already at the lowest role")
		return
	}

	switch claims.Role{
	case "admin":
		if user.UserType == "admin" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Admins cannot demote other admins")
			return
		}
		if user.UserType == "moderator" && req.Role != "user" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Moderators can only be demoted to user by admins")
			return
		}
	case "system_admin":
		if user.UserType == "admin" && req.Role != "moderator" && req.Role != "user" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Admins can only be demoted to moderator or user by system admins")
			return
		}
		if user.UserType == "moderator" && req.Role != "user" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Moderators can only be demoted to user by system admins")
			return
		}
	default:
		utils.WriteError(w, http.StatusForbidden, "Forbidden: Insufficient role")
		return
	}

	// transaction
	tx, err := database.DB.Begin()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback()

	var existingRoleID uuid.UUID
	queryStatement := `
		SELECT role_id
		FROM user_roles
		WHERE user_id = $1
	`
	err = tx.QueryRow(queryStatement, userID).Scan(&existingRoleID)
	if err == sql.ErrNoRows {
		query := `
			INSERT INTO user_roles (user_id, role_id, assigned_by, created_at)
			VALUES ($1, (SELECT id FROM roles WHERE name = $2), $3, $4)
		`
		_, err = tx.Exec(query, userID, req.Role, claims.UserID, time.Now())
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to assign new role")
			return
		}
	} else if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to check existing role")
		return
	} else {
		query := `
			UPDATE user_roles
			SET role_id = (SELECT id FROM roles WHERE name = $1), assigned_by = $2, created_at = $3
			WHERE user_id = $4
		`
		result, err := tx.Exec(query, req.Role, claims.UserID, time.Now(), userID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update user role")
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if rowsAffected == 0 {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update user role")
			return
		}
	}

	queryStatement = `
		UPDATE users
		SET user_type = $1, updated_at = $2
		WHERE id = $3 AND active = TRUE
	`
	result, err := tx.Exec(queryStatement, req.Role, time.Now(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update user type")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if rowsAffected == 0 {
		utils.WriteError(w, http.StatusNotFound, "User not found or already deleted")
		return
	}

	if err := tx.Commit(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to complete role demotion")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User role demoted to %s successfully", req.Role),
	})
}