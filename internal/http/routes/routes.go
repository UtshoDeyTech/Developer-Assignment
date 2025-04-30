package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tahsin005/affpilot-auth/internal/http/handlers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handlers.Health)
	}
	

	return r
}