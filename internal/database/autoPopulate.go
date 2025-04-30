package database

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func seedData(db *gorm.DB) {
	fmt.Println("Seeding data...")

	roles := []models.Role{
		{Name: "system_admin", Description: "Full system access"},
		{Name: "admin", Description: "Admin access"},
		{Name: "moderator", Description: "Can manage content"},
		{Name: "user", Description: "Regular user"},
	}
	for _, role := range roles {
		db.FirstOrCreate(&role, models.Role{Name: role.Name})
	}

	permissions := []models.Permission{
		{Name: "user:read:all", Resource: "user", Action: "read:all", Description: "Read all users"},
		{Name: "user:create:all", Resource: "user", Action: "create:all", Description: "Create users"},
		{Name: "user:update:all", Resource: "user", Action: "update:all", Description: "Update users"},
		{Name: "user:delete:all", Resource: "user", Action: "delete:all", Description: "Delete users"},
		{Name: "user:read:self", Resource: "user", Action: "read:self", Description: "Read own data"},
		{Name: "user:update:self", Resource: "user", Action: "update:self", Description: "Update own data"},
		{Name: "user:delete:self", Resource: "user", Action: "delete:self", Description: "Delete own account"},
		{Name: "role:read", Resource: "role", Action: "read", Description: "Read roles"},
		{Name: "role:create", Resource: "role", Action: "create", Description: "Create roles"},
		{Name: "role:update", Resource: "role", Action: "update", Description: "Update roles"},
		{Name: "role:delete", Resource: "role", Action: "delete", Description: "Delete roles"},
		{Name: "permission:read", Resource: "permission", Action: "read", Description: "Read permissions"},
		{Name: "user:promote:admin", Resource: "user", Action: "promote:admin", Description: "Promote to admin"},
		{Name: "user:promote:moderator", Resource: "user", Action: "promote:moderator", Description: "Promote to moderator"},
		{Name: "user:demote", Resource: "user", Action: "demote", Description: "Demote user"},
	}
	for _, perm := range permissions {
		db.FirstOrCreate(&perm, models.Permission{Name: perm.Name})
	}

	var sysAdmin, adminRole, moderatorRole, userRole models.Role
	db.First(&sysAdmin, "name = ?", "system_admin")
	db.First(&adminRole, "name = ?", "admin")
	db.First(&moderatorRole, "name = ?", "moderator")
	db.First(&userRole, "name = ?", "user")

	var allPerms []models.Permission
	db.Find(&allPerms)
	for _, perm := range allPerms {
		db.FirstOrCreate(&models.RolePermission{RoleID: sysAdmin.ID, PermissionID: perm.ID})
	}

	assignPermissions(db, adminRole.ID, []string{
		"user:read:all", "user:create:all", "user:update:all", "user:delete:all",
		"user:read:self", "user:update:self",
		"role:read", "role:create", "role:update", "role:delete",
		"user:promote:moderator", "user:demote",
	})

	assignPermissions(db, moderatorRole.ID, []string{
		"user:read:all", "user:read:self", "user:update:self", "user:delete:self",
		"user:promote:moderator", "user:demote",
	})

	assignPermissions(db, userRole.ID, []string{
		"user:read:self", "user:update:self", "user:delete:self",
	})

	username := os.Getenv("SYSTEM_ADMIN_USERNAME")
	email := os.Getenv("SYSTEM_ADMIN_EMAIL")
	password := os.Getenv("SYSTEM_ADMIN_PASSWORD")
	usertype := "system_admin"

	if username == "" || email == "" || password == "" {
		log.Println("SYSTEM_ADMIN credentials not set in .env file")
		return
	}

	var existingUser models.User
	result := db.Where("email = ?", email).First(&existingUser)
	if result.Error == gorm.ErrRecordNotFound {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		newUser := models.User{
			Username: username,
			Email:    email,
			Password: string(hashedPassword),
			EmailVerified: true,
			UserType: usertype,
		}
		db.Create(&newUser)

		db.FirstOrCreate(&models.UserRole{
			UserID: newUser.ID,
			RoleID: sysAdmin.ID,
			AssignedBy: newUser.ID,
		})
		log.Println("System admin user created and assigned role.")
	} else {
		log.Println("ℹSystem admin user already exists.")
	}

	log.Println("Seed data inserted successfully")
}

func assignPermissions(db *gorm.DB, roleID uuid.UUID, permNames []string) {
	for _, name := range permNames {
		var perm models.Permission
		if err := db.First(&perm, "name = ?", name).Error; err == nil {
			db.FirstOrCreate(&models.RolePermission{
				RoleID:       roleID,
				PermissionID: perm.ID,
			})
		}
	}
}
