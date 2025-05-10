package database

import (
	"database/sql"
	"log"
)

var DB *sql.DB
func ConnectDB(dbURL string) {
	var err error
	DB, err = sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("DB unreachable:", err)
	}

	log.Println("Database connected successfully")
}