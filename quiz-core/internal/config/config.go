package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBUrl          string
	JWTSecret      string
	InternalSecret string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DBUrl:          getEnv("DB_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", "secret"),
		InternalSecret: getEnv("INTERNAL_SECRET", "internal"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
