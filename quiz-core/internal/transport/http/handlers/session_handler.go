package handlers

import (
	"net/http"

	"github.com/chrpnnkv/SmartBattle/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SessionHandler struct {
	service     *service.SessionService
	quizService *service.QuizService
}

func NewSessionHandler(svc *service.SessionService, quizSvc *service.QuizService) *SessionHandler {
	return &SessionHandler{service: svc, quizService: quizSvc}
}

type CreateSessionReq struct {
	QuizID string `json:"quizId" binding:"required"`
	Mode   string `json:"mode"`
}

type JoinSessionReq struct {
	PIN      string `json:"pin" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
}

// @Summary Создать игровую сессию
// @Tags Sessions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateSessionReq true "ID квиза"
// @Success 201 {object} map[string]interface{}
// @Router /api/sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quizID, err := uuid.Parse(req.QuizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quizId"})
		return
	}
	hostID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
		return
	}

	session, err := h.service.CreateSession(quizID, hostID, req.Mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, buildSessionDTO(h.service, h.quizService, session))
}

// @Summary Присоединиться к игре (Студент)
// @Tags Sessions
// @Accept json
// @Produce json
// @Param request body JoinSessionReq true "PIN код и никнейм"
// @Success 200 {object} map[string]interface{}
// @Router /api/sessions/join [post]
func (h *SessionHandler) JoinSession(c *gin.Context) {
	var req JoinSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.JoinSession(req.PIN)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Получить статус сессии
// @Tags Sessions
// @Produce json
// @Param id path string true "ID Сессии"
// @Success 200 {object} map[string]interface{}
// @Router /api/sessions/{id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	session, err := h.service.GetSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, buildSessionDTO(h.service, h.quizService, session))
}

func (h *SessionHandler) setSessionStatus(c *gin.Context, status string) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	if err := h.service.ChangeStatus(id, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

// @Summary Начать игру (зеркалирует статус в БД; основной поток — через WebSocket).
// @Tags Sessions
// @Security BearerAuth
// @Router /api/sessions/{id}/start [post]
func (h *SessionHandler) StartSession(c *gin.Context) { h.setSessionStatus(c, "active") }

// @Summary Завершить игру (зеркалирует статус в БД; основной поток — через WebSocket).
// @Tags Sessions
// @Security BearerAuth
// @Router /api/sessions/{id}/end [post]
func (h *SessionHandler) EndSession(c *gin.Context) { h.setSessionStatus(c, "finished") }
