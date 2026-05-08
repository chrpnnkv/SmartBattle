package config

import (
	"fmt"
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
	AdminsFile      string
	FrontendURL     string
	UploadsDir      string
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPassword    string
	SMTPFrom        string
}

func (c *Config) Validate() error {
	var missing []string
	if c.DB_DSN == "" {
		missing = append(missing, "DB_DSN")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.XInternalSecret == "" {
		missing = append(missing, "X_INTERNAL_SECRET")
	}
	if c.RealtimeURL == "" {
		missing = append(missing, "REALTIME_URL")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %v", missing)
	}
	return nil
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
		AdminsFile:      getEnv("ADMINS_FILE", "admins.json"),
		FrontendURL:     os.Getenv("FRONTEND_URL"),
		UploadsDir:      getEnv("UPLOADS_DIR", "./uploads"),
		SMTPHost:        os.Getenv("SMTP_HOST"),
		SMTPPort:        os.Getenv("SMTP_PORT"),
		SMTPUser:        os.Getenv("SMTP_USER"),
		SMTPPassword:    os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:        os.Getenv("SMTP_FROM"),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
