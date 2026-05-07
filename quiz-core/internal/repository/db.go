package repository

import (
	"log"
	"os"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// migrationsDir — путь к каталогу с *.sql миграциями.
// Можно переопределить через MIGRATIONS_DIR.
func migrationsDir() string {
	if d := os.Getenv("MIGRATIONS_DIR"); d != "" {
		return d
	}
	return "migrations"
}

func NewPostgresDB(cfg *config.Config) *gorm.DB {
	var (
		db  *gorm.DB
		err error
	)

	// Retry подключения к Postgres
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

	log.Println("Successfully connected to PostgreSQL")

	// Применяем SQL migrations
	if err := RunSQLMigrations(db, migrationsDir()); err != nil {
		log.Fatalf("Failed to apply SQL migrations: %v", err)
	}

	log.Println("SQL migrations applied successfully")

	// Проверяем соединение после миграций
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database ping failed after migrations: %v", err)
	}

	log.Println("Database is ready")

	// AutoMigrate intentionally disabled.
	// Schema is managed exclusively through SQL migrations.
	_ = models.User{}
	_ = models.Quiz{}
	_ = models.Question{}
	_ = models.Option{}
	_ = models.GameSession{}

	return db
}
