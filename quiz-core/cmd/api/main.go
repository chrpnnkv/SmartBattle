package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "quiz-core/docs"
	"quiz-core/internal/config"
	"quiz-core/internal/models"
	"quiz-core/internal/repository"
	"quiz-core/internal/service"
	"quiz-core/internal/transport/rest/handler"
	"quiz-core/internal/transport/rest/middleware"
	"quiz-core/pkg/db"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           Quiz Platform Core API
// @version         1.0
// @description     REST API для управления квизами, пользователями и отчетами.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.email   support@quiz.com

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg := config.LoadConfig()
	database := db.Connect(cfg.DBUrl)

	database.AutoMigrate(&models.User{}, &models.Quiz{}, &models.Question{}, &models.Option{}, &models.GameSession{})

	repo := repository.NewRepository(database)
	svc := service.NewService(repo, cfg)
	h := handler.NewHandler(svc, cfg)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // Ссылка на spec
	))

	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/forgot-password", h.ForgotPassword)
	r.Post("/auth/reset-password", h.ResetPassword)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg.JWTSecret))

		r.Get("/api/me", h.GetMe)
		r.Post("/auth/change-password", h.ChangePassword)

		r.Post("/api/quizzes", h.CreateQuiz)
		r.Put("/api/quizzes/{id}", h.UpdateQuiz)
		r.Get("/api/quizzes", h.ListQuizzes)
		r.Get("/api/quizzes/public", h.ListPublicQuizzes)
		r.Get("/api/quizzes/{id}", h.GetQuiz)
		r.Delete("/api/quizzes/{id}", h.DeleteQuiz)

		r.Get("/api/reports", h.ListReports)
		r.Get("/api/reports/{id}/export", h.ExportReportCSV)
	})

	r.Group(func(r chi.Router) {
		r.Get("/internal/quizzes/{id}", h.InternalGetQuiz)
		r.Post("/internal/reports", h.InternalSaveReport)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited properly")
}
