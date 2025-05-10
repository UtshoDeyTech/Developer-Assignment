package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/tahsin005/affpilot-auth/internal/config"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/routes"
)

func main() {
	fmt.Println("AffPilot Auth Service starting...")
	log.Println("Server initialized")

	config.LoadEnv()
	cfg := config.LoadConfig()
	database.ConnectDB(cfg.DBUrl)
	database.CreateSystemAdminIfNotExists()


	defer database.DB.Close()

	router := routes.RegisterRoutes()
	log.Println("Starting server on port " + cfg.Port)
	log.Fatal(http.ListenAndServe(":" + cfg.Port, router))
}
