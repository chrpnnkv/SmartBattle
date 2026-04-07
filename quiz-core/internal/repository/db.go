package repository

import (
	"log"

	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Quiz{},
		&models.Question{},
		&models.Option{},
		&models.GameSession{},
	)
	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	return db
}
