package config

import "os"

type Config struct {
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	Port        string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ripper:ripper123@localhost:5432/ripper?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}