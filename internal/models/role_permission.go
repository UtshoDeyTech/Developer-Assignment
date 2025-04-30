package models

import (
	"time"

	"github.com/google/uuid"
)

type RolePermission struct {
    RoleID      uuid.UUID `gorm:"type:uuid;primaryKey"`
    PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
    CreatedAt   time.Time `gorm:"autoCreateTime"`
}
