package models

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
    ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    Name        string    `gorm:"size:100;uniqueIndex;not null"`
    Resource    string    `gorm:"size:100;not null"`
    Action      string    `gorm:"size:50;not null"`
    Description string
    CreatedAt   time.Time `gorm:"autoCreateTime"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
