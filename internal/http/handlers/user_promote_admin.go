package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func UserPromoteAdminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if claims.Role != "system_admin" {
		utils.WriteError(w, http.StatusForbidden, "Forbidden: Only system admins can promote to system admin")
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
		utils.WriteError(w, http.StatusForbidden, "Forbidden: You cannot promote yourself")
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

	if user.UserType == "admin" {
		utils.WriteError(w, http.StatusBadRequest, "User is already a admin")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback()

	var existingRoleID uuid.UUID
	queryStatement := `
		SELECT role_id FROM user_roles WHERE user_id = $1
	`
	err = tx.QueryRow(queryStatement, userID).Scan(&existingRoleID)
	if err == sql.ErrNoRows {
		query := `
			INSERT INTO user_roles (user_id, role_id, assigned_by, created_at)
			VALUES ($1, (SELECT id FROM roles WHERE name = $2), $3, $4)
		`
		_, err = tx.Exec(query, userID, "admin", claims.UserID, time.Now())
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to assign admin role")
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
		result, err := tx.Exec(query, "admin", claims.UserID, time.Now(), userID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update user role to admin")
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if rowsAffected == 0 {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update user role to admin")
			return
		}
	}

	query := `
		UPDATE users
		SET user_type = $1, updated_at = $2
		WHERE id = $3 AND active = TRUE
	`
	result, err := tx.Exec(query, "admin", time.Now(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update user type to admin")
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
		utils.WriteError(w, http.StatusInternalServerError, "Failed to complete promotion")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User role changed to admin successfully",
	})
}