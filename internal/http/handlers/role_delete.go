package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

type RoleDeleteResponse struct {
	Message string `json:"message"`
}

func RoleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	roleIDStr := vars["role_id"]

	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid role_id format")
		return
	}

	log.Printf("User %s (role: %s) deleting role %s", claims.UserID, claims.Role, roleIDStr)

	query := `
		DELETE FROM roles
		WHERE id = $1
	`
	result, err := database.DB.Exec(query, roleID)
	if err != nil {
		log.Printf("Failed to delete role: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to check rows affected: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "Failed to verify role deletion")
		return
	}
	if rowsAffected == 0 {
		utils.WriteError(w, http.StatusNotFound, "Role not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RoleDeleteResponse{
		Message: "Role deleted successfully",
	})
}