package handlers

import (
	"net/http"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QuizHandler struct {
	service *service.QuizService
}

func NewQuizHandler(service *service.QuizService) *QuizHandler {
	return &QuizHandler{service: service}
}

// @Summary Создание Квиза
// @Tags Quizzes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.Quiz true "Структура квиза"
// @Success 201 {object} models.Quiz
// @Router /api/quizzes [post]
func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var quiz models.Quiz
	if err := c.ShouldBindJSON(&quiz); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	quiz.TeacherID = teacherID

	if err := h.service.CreateQuiz(&quiz); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, quiz)
}

// @Summary Список квизов преподавателя
// @Tags Quizzes
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Quiz
// @Router /api/quizzes [get]
func (h *QuizHandler) GetQuizzes(c *gin.Context) {
	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	quizzes, err := h.service.GetTeacherQuizzes(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quizzes)
}

// @Summary Получить квиз по ID
// @Tags Quizzes
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID Квиза"
// @Success 200 {object} models.Quiz
// @Router /api/quizzes/{id} [get]
func (h *QuizHandler) GetQuizByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	quiz, err := h.service.GetQuizByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quiz not found"})
		return
	}
	c.JSON(http.StatusOK, quiz)
}

// @Summary Обновление Квиза
// @Tags Quizzes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID Квиза"
// @Param request body models.Quiz true "Новая структура квиза"
// @Success 200 {object} models.Quiz
// @Router /api/quizzes/{id} [put]
func (h *QuizHandler) UpdateQuiz(c *gin.Context) {
	var quiz models.Quiz
	if err := c.ShouldBindJSON(&quiz); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	quiz.ID = id
	quiz.TeacherID, _ = uuid.Parse(c.GetString("user_id"))

	if err := h.service.UpdateQuiz(&quiz); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quiz)
}

// @Summary Удалить квиз
// @Tags Quizzes
// @Security BearerAuth
// @Param id path string true "ID Квиза"
// @Success 204
// @Router /api/quizzes/{id} [delete]
func (h *QuizHandler) DeleteQuiz(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	if err := h.service.DeleteQuiz(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
