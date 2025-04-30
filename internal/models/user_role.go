package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
    UserID     uuid.UUID `gorm:"type:uuid;primaryKey"`
    RoleID     uuid.UUID `gorm:"type:uuid;primaryKey"`
    AssignedBy uuid.UUID `gorm:"type:uuid;not null"`
    CreatedAt  time.Time `gorm:"autoCreateTime"`
}
