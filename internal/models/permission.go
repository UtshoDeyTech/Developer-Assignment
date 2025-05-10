package models

import "github.com/google/uuid"

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description,omitempty"`
}

type PermissionListResponse struct {
	Message     string       `json:"message"`
	Permissions []Permission `json:"permissions"`
}

type PermissionDetailsResponse struct {
	Message    string     `json:"message"`
	Permission Permission `json:"permission"`
}
