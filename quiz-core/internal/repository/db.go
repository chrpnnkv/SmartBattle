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

// migrationsDir — путь к каталогу с *.sql миграциями. Переопределяется
// переменной окружения MIGRATIONS_DIR (полезно для тестов и нестандартных деплоев).
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

	// Сначала применяем версионированные SQL-миграции (см. migrations/).
	// Они каноническая правда о схеме; AutoMigrate ниже — safety-net на случай
	// расхождения моделей и SQL во время разработки.
	if err := RunSQLMigrations(db, migrationsDir()); err != nil {
		log.Fatalf("Failed to apply SQL migrations: %v", err)
	}

	// AutoMigrate с ретраями: если БД только поднялась — даём ей шанс.
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
		log.Fatalf("Failed to run AutoMigrate after retries: %v", err)
	}

	return db
}
