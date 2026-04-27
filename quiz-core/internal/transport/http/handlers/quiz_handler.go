package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/chrpnnkv/SmartBattle/internal/models"
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

type OptionInput struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
	Color     string `json:"color"`
}

type QuestionInput struct {
	ID                 string        `json:"id"`
	Type               string        `json:"type"`
	Text               string        `json:"text"`
	ImageURL           string        `json:"imageUrl"`
	TimeLimitSeconds   int           `json:"timeLimitSeconds"`
	Score              int           `json:"score"`
	Options            []OptionInput `json:"options"`
	CorrectTextAnswers []string      `json:"correctTextAnswers"`
}

type QuizInput struct {
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Mode        string                 `json:"mode"`
	Settings    map[string]interface{} `json:"settings"`
	Questions   []QuestionInput        `json:"questions"`
}

func mapQuizInputToModel(input QuizInput, teacherID uuid.UUID) *models.Quiz {
	quiz := &models.Quiz{
		TeacherID:   teacherID,
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		Mode:        input.Mode,
	}

	settingsBytes, _ := json.Marshal(input.Settings)
	quiz.Settings = datatypes.JSON(settingsBytes)

	for i, qIn := range input.Questions {
		qID, err := uuid.Parse(qIn.ID)
		if err != nil {
			qID = uuid.New()
		}

		correctAnswersBytes, _ := json.Marshal(qIn.CorrectTextAnswers)

		q := models.Question{
			ID:                 qID,
			Type:               qIn.Type,
			Text:               qIn.Text,
			ImageURL:           qIn.ImageURL,
			TimerSec:           qIn.TimeLimitSeconds,
			Score:              qIn.Score,
			Order:              i,
			CorrectTextAnswers: datatypes.JSON(correctAnswersBytes),
		}

		for _, oIn := range qIn.Options {
			oID, err := uuid.Parse(oIn.ID)
			if err != nil {
				oID = uuid.New()
			}
			q.Options = append(q.Options, models.Option{
				ID:        oID,
				Text:      oIn.Text,
				IsCorrect: oIn.IsCorrect,
				Color:     oIn.Color,
			})
		}
		quiz.Questions = append(quiz.Questions, q)
	}
	return quiz
}

func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var input QuizInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	quiz := mapQuizInputToModel(input, teacherID)

	if err := h.service.CreateQuiz(quiz); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, quiz)
}

func (h *QuizHandler) GetQuizzes(c *gin.Context) {
	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	quizzes, err := h.service.GetTeacherQuizzes(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quizzes)
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
	c.JSON(http.StatusOK, quiz)
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

	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	quiz := mapQuizInputToModel(input, teacherID)
	quiz.ID = id

	if err := h.service.UpdateQuiz(quiz); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quiz)
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
