package config

import (
	"fmt"
	"os"
	"ripper-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

var DB *gorm.DB

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() *Config {
	// Try to get DATABASE_URL first (for docker-compose)
	databaseURL := os.Getenv("DATABASE_URL")

	// If not set, build from individual components (for Kubernetes)
	if databaseURL == "" {
		host := getEnv("DATABASE_HOST", "localhost")
		port := getEnv("DATABASE_PORT", "5432")
		user := getEnv("DATABASE_USER", "ripper")
		password := getEnv("DATABASE_PASSWORD", "ripper123")
		dbname := getEnv("DATABASE_NAME", "ripper")
		sslmode := getEnv("DATABASE_SSLMODE", "disable")

		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
	}

	return &Config{
		DatabaseURL: databaseURL,
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		Port:        getEnv("PORT", "8080"),
	}
}

func ConnectDB() {
	config := Load()
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	db.AutoMigrate(
		&models.User{},
		&models.TwitterAccount{},
		&models.WhatsAppAccount{},
		&models.ApiCallLog{},
		&models.MessageLog{}, // New: Message history logging
	)
	DB = db
}
