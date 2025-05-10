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

func PermissionDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	permissionIDStr := vars["permission_id"]

	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid permission_id format")
		return
	}

	log.Printf("User %s (role: %s) retrieving details for permission %s", claims.UserID, claims.Role, permissionIDStr)

	var p models.Permission
	var description sql.NullString
	query := `
		SELECT id, name, resource, action, description
		FROM permissions
		WHERE id = $1
	`
	err = database.DB.QueryRow(query, permissionID).Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &description)
	if err == sql.ErrNoRows {
		utils.WriteError(w, http.StatusNotFound, "Permission not found")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve permission details")
		return
	}
	p.Description = description.String

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.PermissionDetailsResponse{
		Message:    "Permission details retrieved successfully",
		Permission: p,
	})
}