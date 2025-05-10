package models

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

type RoleCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type RoleCreateResponse struct {
	Message string `json:"message"`
	Role    Role   `json:"role"`
}

type RoleUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleUpdateResponse struct {
	Message string `json:"message"`
	Role    Role   `json:"role"`
}

type RoleDemoteRequest struct {
	Role string `json:"role"`
}

type RoleChangeRequest struct {
	Role string `json:"role"`
}

type RoleListResponse struct {
	Message string `json:"message"`
	Roles   []Role `json:"roles"`
}

type RoleDetailsResponse struct {
	Message string `json:"message"`
	Role    Role   `json:"role"`
}
