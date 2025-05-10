package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func PermissionListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	log.Printf("User %s (role: %s) retrieving permissions list", claims.UserID, claims.Role)

	var permissions []models.Permission
	query := `
		SELECT id, name, resource, action, description
		FROM permissions
		ORDER BY name
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve permissions")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Permission
		var description sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &description); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to parse permissions")
			return
		}
		p.Description = description.String
		permissions = append(permissions, p)
	}

	if err := rows.Err(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Error processing permissions")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.PermissionListResponse{
		Message:     "Permissions list retrieved successfully",
		Permissions: permissions,
	})
}