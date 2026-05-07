package http

import (
	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/transport/http/handlers"
	"github.com/chrpnnkv/SmartBattle/internal/transport/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/chrpnnkv/SmartBattle/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	cfg *config.Config,
	authH *handlers.AuthHandler,
	quizH *handlers.QuizHandler,
	reportH *handlers.ReportHandler,
	sessionH *handlers.SessionHandler,
	uploadH *handlers.UploadHandler,
) *gin.Engine {
	r := gin.Default()

	configCORS := cors.DefaultConfig()
	configCORS.AllowAllOrigins = true
	configCORS.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Internal-Secret"}
	r.Use(cors.New(configCORS))

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Статическая раздача загруженных изображений
	r.Static("/uploads", cfg.UploadsDir)

	r.POST("/api/sessions/join", sessionH.JoinSession)
	r.GET("/api/sessions/:id", sessionH.GetSession)

	// Публичный каталог квизов — без авторизации.
	// Отдельный путь, чтобы не конфликтовать с GET /api/quizzes/:id (Gin radix-tree).
	r.GET("/api/public/quizzes", quizH.GetPublicQuizzes)

	auth := r.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/forgot-password", authH.ForgotPassword)
		auth.POST("/reset-password", authH.ResetPassword)
		auth.POST("/change-password", middleware.AuthGuard(cfg), authH.ChangePassword)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthGuard(cfg))
	{
		api.GET("/me", authH.GetMe)

		api.GET("/quizzes", quizH.GetQuizzes)
		api.POST("/quizzes", quizH.CreateQuiz)
		api.GET("/quizzes/:id", quizH.GetQuizByID)
		api.PUT("/quizzes/:id", quizH.UpdateQuiz)
		api.DELETE("/quizzes/:id", quizH.DeleteQuiz)

		api.POST("/uploads/image", uploadH.UploadImage)

		api.GET("/reports", reportH.GetReports)
		api.GET("/reports/:id", reportH.GetReportByID)
		api.GET("/reports/:id/export", reportH.ExportCSV)

		api.POST("/sessions", sessionH.CreateSession)
		api.POST("/sessions/:id/start", sessionH.StartSession)
		api.POST("/sessions/:id/end", sessionH.EndSession)
		// Управление переходом к следующему вопросу и приёмом ответов
		// идёт через WebSocket Realtime — отдельных HTTP-endpoint'ов в Core нет.
	}

	internal := r.Group("/internal")
	internal.Use(middleware.InternalGuard(cfg))
	{
		internal.GET("/quizzes/:id", quizH.GetQuizByID)
		internal.POST("/reports", reportH.SaveInternalReport)
	}

	return r
}
