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

func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, http.StatusForbidden, "Forbidden: You cannot delete your own account")
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

	// preventing deletion of system admins
	if user.UserType == "system_admin" {
		utils.WriteError(w, http.StatusForbidden, "Forbidden: System admins cannot be deleted")
		return
	}

	switch claims.Role {
	case "moderator":
		// moderators can only delete users who requested deletion and are not admins
		if !user.DeletionRequested {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: User has not requested deletion")
			return
		}
		if user.UserType == "admin" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Moderators cannot delete admins")
			return
		}
	case "admin":
		// admins cannot delete other admins
		if user.UserType == "admin" {
			utils.WriteError(w, http.StatusForbidden, "Forbidden: Admins cannot delete other admins")
			return
		}
	case "system_admin":
		// system admins can delete anyone except system admins (already checked)
		log.Printf("System admin %s authorized to delete user %s", claims.UserID, userIDStr)
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

	// soft deletion
	queryStatement := `
		UPDATE users
		SET active = FALSE, updated_at = $1
		WHERE id = $2 AND active = TRUE
	`
	result, err := tx.Exec(queryStatement, time.Now(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to delete user")
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

	queryStatement = `
		DELETE FROM user_roles
		WHERE user_id = $1
	`
	_, err = tx.Exec(queryStatement, userID)
	if err != nil {
		log.Printf("Warning: Failed to clean up user_roles for user %s: %v", userID, err)
	}

	if err := tx.Commit(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to complete deletion")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
}
