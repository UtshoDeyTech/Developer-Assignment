package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tahsin005/affpilot-auth/internal/http/handlers"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handlers.Health)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterHandler(db))
		}
	}

	return r
}
