package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func RoleDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("User %s (role: %s) retrieving details for role %s", claims.UserID, claims.Role, roleIDStr)

	var role models.Role
	var description sql.NullString
	query := `
		SELECT id, name, description
		FROM roles
		WHERE id = $1
	`
	err = database.DB.QueryRow(query, roleID).Scan(&role.ID, &role.Name, &description)
	if err == sql.ErrNoRows {
		utils.WriteError(w, http.StatusNotFound, "Role not found")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve role details")
		return
	}
	role.Description = description.String

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.RoleDetailsResponse{
		Message: "Role details retrieved successfully",
		Role:    role,
	})
}