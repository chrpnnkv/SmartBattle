package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/admins"
	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/realtime"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/chrpnnkv/SmartBattle/internal/service"
	transportHttp "github.com/chrpnnkv/SmartBattle/internal/transport/http"
	"github.com/chrpnnkv/SmartBattle/internal/transport/http/handlers"
)

// @title           Smart Battle API
// @version         1.0
// @description     Backend Core для платформы академических квизов
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey InternalSecretAuth
// @in header
// @name X-Internal-Secret
func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("конфигурация невалидна: %v", err)
	}
	db := repository.NewPostgresDB(cfg)

	userRepo := repository.NewUserRepository(db)
	quizRepo := repository.NewQuizRepository(db)
	reportRepo := repository.NewReportRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Список администраторов: загружается из statically-managed JSON-файла.
	// Отсутствие файла — допустимый режим: никто не считается администратором.
	adminList, err := admins.LoadFromFile(cfg.AdminsFile)
	if err != nil {
		log.Fatalf("не удалось загрузить список администраторов из %s: %v", cfg.AdminsFile, err)
	}

	emailSvc := service.NewSMTPEmailService(cfg)
	authService := service.NewAuthService(userRepo, cfg, adminList, emailSvc)
	quizService := service.NewQuizService(quizRepo)
	reportService := service.NewReportService(reportRepo)

	// HTTP-клиент к realtime — единственная точка интеграции, скрытая за интерфейсом.
	// В тестах SessionService можно подменить на фейк (см. internal/realtime/Client).
	rtClient := realtime.NewHTTPClient(cfg.RealtimeURL, cfg.JWTSecret)
	sessionService := service.NewSessionService(sessionRepo, quizRepo, rtClient)

	authHandler := handlers.NewAuthHandler(authService)
	quizHandler := handlers.NewQuizHandler(quizService)
	reportHandler := handlers.NewReportHandler(reportService, quizService)
	sessionHandler := handlers.NewSessionHandler(sessionService, quizService)

	router := transportHttp.SetupRouter(cfg, authHandler, quizHandler, reportHandler, sessionHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server starting on port %s", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
