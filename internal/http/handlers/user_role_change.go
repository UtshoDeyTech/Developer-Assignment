package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

func UserRoleChangeHandler(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, http.StatusForbidden, "Forbidden: You cannot change your own role")
		return
	}

	var req models.RoleChangeRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	validRoles := map[string]bool{
		"user": true,
		"moderator": true,
		"admin": true,
		"system_admin": true,
	}

	if !validRoles[req.Role] {
		utils.WriteError(w, http.StatusBadRequest, "Invalid role: only accepts user, moderator, admin & system_admin")
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

	// preventing role change for system admin
	if user.UserType == "system_admin" {
		utils.WriteError(w, http.StatusForbidden, "Forbidden: System admins' roles cannot be changed")
		return
	}

	// role change rules
	switch claims.Role {
	case "admin":
		if req.Role == "system_admin" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Admins cannot assign system_admin role")
			return
		}
		if user.UserType == "admin" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Admins cannot change other admins' roles")
			return
		}
		if user.UserType == "moderator" && req.Role != "user" && req.Role != "admin" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Moderators can only be changed to user or admin")
			return
		}
		if user.UserType == "user" && req.Role != "moderator" && req.Role != "admin" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Users can only be changed to moderator or admin")
			return
		}
	case "system_admin":
		log.Printf("System admin %s authorized to change role of user %s to %s", claims.UserID, userIDStr, req.Role)
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

	queryStatement := `
		SELECT role_id 
		FROM user_roles
		WHERE user_id = $1
	`
	var existingRoleID uuid.UUID
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

	query := `
		UPDATE users
		SET user_type = $1, updated_at = $2
		WHERE id = $3 AND active = TRUE
	`
	result, err := tx.Exec(query, req.Role, time.Now(), userID)
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
		utils.WriteError(w, http.StatusInternalServerError, "Failed to complete role change")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User role changed to %s successfully", req.Role),
	})
}