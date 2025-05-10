package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl string
	Port string
	JWT_SECRET string
}

type DBConfig struct {
	DBHost string
    DBPort string
    DBUser string
    DBPassword string
    DBName string
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}

func LoadConfig() *Config {
	dbConfig := DBConfig {
		DBHost:     os.Getenv("DB_HOST"),
        DBPort:     os.Getenv("DB_PORT"),
        DBUser:     os.Getenv("DB_USER"),
        DBPassword: os.Getenv("DB_PASSWORD"),
        DBName:     os.Getenv("DB_NAME"),
	}
	dbURL := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbConfig.DBHost, dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBName, dbConfig.DBPort,
	)
	port := os.Getenv("SERVER_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	if port == "" {
		port = "8080"
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	return &Config{
		DBUrl: dbURL,
		JWT_SECRET: jwtSecret,
		Port:  port,
	}
}