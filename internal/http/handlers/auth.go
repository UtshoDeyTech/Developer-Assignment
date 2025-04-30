package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tahsin005/affpilot-auth/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest

		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		var existingUser models.User
		result := db.Where("username = ?", req.Username).First(&existingUser)
		if result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Username already exists",
			})
			return
		}

		result = db.Where("email = ?", strings.ToLower(req.Email)).First(&existingUser) 
		if result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already registered",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process registration",
			})
			return
		}

		verificationToken := uuid.New().String()
		tokenExpiry := time.Now().Add(24 * time.Hour)

		user := models.User{
			ID:                uuid.New(),
			Username:          req.Username,
			Email:             strings.ToLower(req.Email),
			Password:          string(hashedPassword),
			FirstName:         req.FirstName,
			LastName:          req.LastName,
			EmailVerified:     false,
			UserType:          "user",
			VerificationToken: verificationToken,
			TokenExpiry:       &tokenExpiry,
			DeletionRequested: false,
			Active:            true,
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to register user",
			})
			return
		}

		var userRoleModel models.Role
		var systemAdminRoleModel models.Role

		userRoleModelResult := db.Where("name = ?", "user").First(&userRoleModel)
		systemAdminRoleModelResult := db.Where("name = ?", "system_admin").First(&systemAdminRoleModel)

		if userRoleModelResult.Error != nil || systemAdminRoleModelResult.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get a role",
			})
			return
		}

		userRole := models.UserRole{
			UserID:     user.ID,
			RoleID:     userRoleModel.ID,
			AssignedBy: systemAdminRoleModel.ID,
		}

		if err := db.Create(&userRole).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to assign role",
			})
			return
		}

		// TODO: send email with verification link

		user.Password = ""

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully.",
			"user":    user,
		})
	}
}
