package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"github.com/tahsin005/affpilot-auth/internal/utils"
)

func RoleCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.RoleCreateRequest
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

	log.Printf("User %s (role: %s) creating role %s", claims.UserID, claims.Role, req.Name)

	roleID := uuid.New()

	query := `
		INSERT INTO roles (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, name, description
	`
	var role models.Role
	var description sql.NullString
	err := database.DB.QueryRow(query, roleID, req.Name, req.Description).Scan(&role.ID, &role.Name, &description)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			utils.WriteError(w, http.StatusConflict, "Role name already exists")
			return
		}
		log.Printf("Failed to create role: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create role")
		return
	}
	role.Description = description.String

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.RoleCreateResponse{
		Message: "Role created successfully",
		Role:    role,
	})
}