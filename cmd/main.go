package main

import (
	"log"
	"os"

	"github.com/tahsin005/affpilot-auth/internal/config"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/routes"
)

func main() {
	cfg := config.LoadConfig()
	err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Initialization error: %v", err)
	}

	r := routes.SetupRouter()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
