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

func RoleListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	log.Printf("User %s (role: %s) retrieving roles list", claims.UserID, claims.Role)

	var roles []models.Role
	query := `
		SELECT id, name, description
		FROM roles
		ORDER BY name
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve roles")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Role
		var description sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &description); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to parse roles")
			return
		}
		r.Description = description.String
		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Error processing roles")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.RoleListResponse{
		Message: "Roles list retrieved successfully",
		Roles:   roles,
	})
}