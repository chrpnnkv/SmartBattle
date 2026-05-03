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
	if input.Questions != nil {
		teacherID, parseErr := uuid.Parse(c.GetString("user_id"))
		if parseErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		fresh := mapQuizInputToModel(input, teacherID)
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

	if err := h.service.DeleteQuiz(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
