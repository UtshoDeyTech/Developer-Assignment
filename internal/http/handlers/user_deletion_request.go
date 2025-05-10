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

func UserDeletionRequestHandler(w http.ResponseWriter, r *http.Request) {
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

	if userIDStr != claims.UserID {
		utils.WriteError(w, http.StatusForbidden, "Forbidden: You can only request deletion for your own account")
		return
	}

	user, err := utils.GetUserByID(userID)
	if err == sql.ErrNoRows {
		utils.WriteError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if user.DeletionRequested {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "You have already requested deletion; please wait for the update",
		})
		return
	}

	queryStatement := `
		UPDATE users
		SET deletion_requested = TRUE, updated_at = $1
		WHERE id = $2 AND active = TRUE
	`
	result, err := database.DB.Exec(queryStatement, time.Now(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to request account deletion")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if rowsAffected == 0 {
		utils.WriteError(w, http.StatusNotFound, "User not found or already inactive")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Your deletion request has been sent successfully",
	})
}
