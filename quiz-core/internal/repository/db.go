package repository

import (
	"log"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(cfg *config.Config) *gorm.DB {
	var (
		db  *gorm.DB
		err error
	)

	// Устойчивое подключение к Postgres с retry (важно для docker startup race)
	for attempt := 1; attempt <= 20; attempt++ {
		db, err = gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil {
				pingErr := sqlDB.Ping()
				if pingErr == nil {
					break
				}
				err = pingErr
			} else {
				err = dbErr
			}
		}

		log.Printf("Postgres connection attempt %d/20 failed: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}

	// Миграции тоже ретраим, если БД только поднялась и еще не готова к DDL
	for attempt := 1; attempt <= 10; attempt++ {
		err = db.AutoMigrate(
			&models.User{},
			&models.Quiz{},
			&models.Question{},
			&models.Option{},
			&models.GameSession{},
		)
		if err == nil {
			break
		}
		log.Printf("AutoMigrate attempt %d/10 failed: %v", attempt, err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to run database migrations after retries: %v", err)
	}

	return db
}
