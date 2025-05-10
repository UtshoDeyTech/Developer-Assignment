package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func RoleUpdateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req models.RoleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.WriteError(w, http.StatusBadRequest, "Role name is required")
		return
	}
	if len(req.Name) > 50 {
		utils.WriteError(w, http.StatusBadRequest, "Role name must be 50 characters or less")
		return
	}

	if strings.TrimSpace(req.Description) == "" {
		utils.WriteError(w, http.StatusBadRequest, "Role description is required")
		return
	}

	log.Printf("User %s (role: %s) updating role %s with name %s", claims.UserID, claims.Role, roleIDStr, req.Name)

	queryStatement := `
		UPDATE roles
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description
	`

	var role models.Role
	var description sql.NullString
	err = database.DB.QueryRow(queryStatement, req.Name, req.Description, roleID).Scan(&role.ID, &role.Name, &description)
	if err == sql.ErrNoRows {
		utils.WriteError(w, http.StatusNotFound, "Role not found")
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			utils.WriteError(w, http.StatusConflict, "Role name already exists")
			return
		}
		log.Printf("Failed to update role: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	role.Description = description.String

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.RoleUpdateResponse{
		Message: "Role updated successfully",
		Role:    role,
	})
}