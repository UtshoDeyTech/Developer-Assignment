package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func UserPromoteModeratorHandler(w http.ResponseWriter, r *http.Request) {
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

	switch user.UserType {
	case "moderator":
		utils.WriteError(w, http.StatusBadRequest, "User is already a moderator")
		return
	case "admin", "system_admin":
		utils.WriteError(w, http.StatusForbidden, "Forbidden: Cannot promote admin or system admin to moderator")
		return
	case "user":
		log.Println("Eligible for promotion")
	default:
		utils.WriteError(w, http.StatusBadRequest, "Invalid user role")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback()

	var existingRoleID uuid.UUID
	err = tx.QueryRow("SELECT role_id FROM user_roles WHERE user_id = $1", userID).Scan(&existingRoleID)
	if err == sql.ErrNoRows {
		query := `
			INSERT INTO user_roles (user_id, role_id, assigned_by, created_at)
			VALUES ($1, (SELECT id FROM roles WHERE name = $2), $3, $4)
		`
		_, err = tx.Exec(query, userID, "moderator", claims.UserID, time.Now())
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to assign moderator role")
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
		result, err := tx.Exec(query, "moderator", claims.UserID, time.Now(), userID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update user role to moderator")
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if rowsAffected == 0 {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update user role to moderator")
			return
		}
	}

	query := `
		UPDATE users
		SET user_type = $1, updated_at = $2
		WHERE id = $3 AND active = TRUE
	`
	result, err := tx.Exec(query, "moderator", time.Now(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update user type to moderator")
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
		"message": "User role changed to moderator successfully",
	})
}