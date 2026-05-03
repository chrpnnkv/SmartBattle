package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DB_DSN          string
	JWTSecret       string
	XInternalSecret string
	RealtimeURL     string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:            port,
		DB_DSN:          os.Getenv("DB_DSN"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		XInternalSecret: os.Getenv("X_INTERNAL_SECRET"),
		RealtimeURL:     getEnv("REALTIME_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
