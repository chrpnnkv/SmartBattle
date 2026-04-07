package http

import (
	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/transport/http/handlers"
	"github.com/chrpnnkv/SmartBattle/internal/transport/middleware"
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
) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Подключение Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

		api.GET("/reports", reportH.GetReports)
		api.GET("/reports/:id/export", reportH.ExportCSV)
	}

	internal := r.Group("/internal")
	internal.Use(middleware.InternalGuard(cfg))
	{
		internal.GET("/quizzes/:id", quizH.GetQuizByID)
		internal.POST("/reports", reportH.SaveInternalReport)
	}

	return r
}
