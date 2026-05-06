package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/chrpnnkv/SmartBattle/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type QuizHandler struct {
	service *service.QuizService
}

func NewQuizHandler(service *service.QuizService) *QuizHandler {
	return &QuizHandler{service: service}
}

func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var createInput CreateQuizInput
	if err := c.ShouldBindJSON(&createInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacherID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
		return
	}
	input := QuizInput{
		Title:       createInput.Title,
		Description: createInput.Description,
		Status:      createInput.Status,
		Mode:        createInput.Mode,
		Settings:    createInput.Settings,
		Questions:   createInput.Questions,
	}
	quiz := mapQuizInputToModel(input, teacherID)

	if err := h.service.CreateQuiz(quiz); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Перечитываем с предзагрузкой Questions/Options, иначе DTO будет пустым.
	if fresh, err := h.service.GetQuizByID(quiz.ID); err == nil {
		c.JSON(http.StatusCreated, mapQuizToDTO(fresh))
		return
	}
	c.JSON(http.StatusCreated, mapQuizToDTO(quiz))
}

func (h *QuizHandler) GetQuizzes(c *gin.Context) {
	// Администратор видит все квизы платформы; обычный преподаватель — только свои.
	if c.GetString("role") == "admin" {
		quizzes, err := h.service.GetAllQuizzes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapQuizzesToDTO(quizzes))
		return
	}

	teacherID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
		return
	}
	quizzes, err := h.service.GetTeacherQuizzes(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapQuizzesToDTO(quizzes))
}

func (h *QuizHandler) GetPublicQuizzes(c *gin.Context) {
	quizzes, err := h.service.GetPublicQuizzes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapQuizzesToDTO(quizzes))
}

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
	c.JSON(http.StatusOK, mapQuizToDTO(quiz))
}

func (h *QuizHandler) UpdateQuiz(c *gin.Context) {
	var input QuizInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	existing, err := h.service.GetQuizByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quiz not found"})
		return
	}

	// Доступ к редактированию: либо автор квиза, либо администратор.
	if !canManageQuiz(c, existing.TeacherID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: not the quiz owner"})
		return
	}

	// Merge: применяем только присланные поля, остальное сохраняем.
	if input.Title != "" {
		existing.Title = input.Title
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if input.Status != "" {
		existing.Status = input.Status
	}
	if input.Mode != "" {
		existing.Mode = input.Mode
	}
	if input.Settings != nil {
		settingsBytes, _ := json.Marshal(input.Settings)
		existing.Settings = datatypes.JSON(settingsBytes)
	}
	// Questions перезаписываем, только если поле явно прислано (не nil).
	// При редактировании администратором используем teacherID самого квиза, а не
	// идентификатор администратора — это сохраняет авторство в БД.
	if input.Questions != nil {
		fresh := mapQuizInputToModel(input, existing.TeacherID)
		existing.Questions = fresh.Questions
	}

	if err := h.service.UpdateQuiz(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if fresh, err := h.service.GetQuizByID(existing.ID); err == nil {
		c.JSON(http.StatusOK, mapQuizToDTO(fresh))
		return
	}
	c.JSON(http.StatusOK, mapQuizToDTO(existing))
}

func (h *QuizHandler) DeleteQuiz(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	// Удалять может либо автор, либо администратор.
	existing, err := h.service.GetQuizByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quiz not found"})
		return
	}
	if !canManageQuiz(c, existing.TeacherID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: not the quiz owner"})
		return
	}

	if err := h.service.DeleteQuiz(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// canManageQuiz возвращает true, если из контекста Gin следует, что вызывающий
// либо является автором квиза с teacherID, либо имеет роль администратора.
// Хелпер используется в UpdateQuiz и DeleteQuiz; вынесен сюда, чтобы логика
// контроля доступа была локализована и согласована.
func canManageQuiz(c *gin.Context, teacherID uuid.UUID) bool {
	if c.GetString("role") == "admin" {
		return true
	}
	currentID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		return false
	}
	return currentID == teacherID
}
