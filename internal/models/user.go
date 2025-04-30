package models

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID                 uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    Username           string    `gorm:"uniqueIndex;size:50;not null"`
    Email              string    `gorm:"uniqueIndex;size:100;not null"`
    Password       string    `gorm:"size:100;not null"`
    FirstName          string    `gorm:"size:50"`
    LastName           string    `gorm:"size:50"`
    EmailVerified      bool      `gorm:"default:false"`
    UserType           string    `gorm:"size:20;not null"`
    VerificationToken  string    `gorm:"size:100"`
    TokenExpiry        *time.Time
    DeletionRequested  bool      `gorm:"default:false"`
    Active             bool      `gorm:"default:true"`
    CreatedAt          time.Time `gorm:"autoCreateTime"`
    UpdatedAt          time.Time `gorm:"autoUpdateTime"`

    Roles []Role `gorm:"many2many:user_roles;"`
}
