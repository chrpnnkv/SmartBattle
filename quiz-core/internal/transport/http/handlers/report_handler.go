package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportHandler struct {
	service *service.ReportService
}

func NewReportHandler(service *service.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

// DTO-структура, которую ждет фронтенд
type GameReportDTO struct {
	ID             string                 `json:"id"`
	QuizID         string                 `json:"quizId"`
	HostID         string                 `json:"hostId"`
	PIN            string                 `json:"pin"`
	Status         string                 `json:"status"`
	StartedAt      time.Time              `json:"startedAt"`
	FinishedAt     *time.Time             `json:"finishedAt"`
	ReportSnapshot map[string]interface{} `json:"reportSnapshot"`
}

func mapToDTO(s models.GameSession) GameReportDTO {
	dto := GameReportDTO{
		ID:         s.ID.String(),
		QuizID:     s.QuizID.String(),
		HostID:     s.HostID.String(),
		PIN:        s.PIN,
		Status:     s.Status,
		StartedAt:  s.StartedAt,
		FinishedAt: s.FinishedAt,
	}
	// Парсим raw bytes из БД в красивый JSON-объект
	if len(s.ReportSnapshot) > 0 {
		var snap map[string]interface{}
		json.Unmarshal(s.ReportSnapshot, &snap)
		dto.ReportSnapshot = snap
	} else {
		dto.ReportSnapshot = make(map[string]interface{})
	}
	return dto
}

// @Summary INTERNAL: Сохранить результаты игры
// @Tags Internal
// @Security InternalSecretAuth
// @Accept json
// @Produce json
// @Param request body models.GameSession true "Снапшот сессии"
// @Success 201
// @Router /internal/reports [post]
func (h *ReportHandler) SaveInternalReport(c *gin.Context) {
	var session models.GameSession
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SaveSessionReport(&session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

// @Summary История игр преподавателя
// @Tags Analytics
// @Security BearerAuth
// @Produce json
// @Success 200 {array} GameReportDTO
// @Router /api/reports [get]
func (h *ReportHandler) GetReports(c *gin.Context) {
	hostID, _ := uuid.Parse(c.GetString("user_id"))
	sessions, err := h.service.GetTeacherReports(hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var reports []GameReportDTO
	for _, s := range sessions {
		reports = append(reports, mapToDTO(s))
	}

	if reports == nil {
		reports = []GameReportDTO{} // Возвращаем [] вместо null для фронтенда
	}

	c.JSON(http.StatusOK, reports)
}

// @Summary Детальный отчет по игре
// @Tags Analytics
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID Сессии"
// @Success 200 {object} GameReportDTO
// @Router /api/reports/{id} [get]
func (h *ReportHandler) GetReportByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	session, err := h.service.GetReportByID(id)
	if err != nil || session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	c.JSON(http.StatusOK, mapToDTO(*session))
}

// @Summary Скачать CSV отчет
// @Tags Analytics
// @Security BearerAuth
// @Produce text/csv
// @Param id path string true "ID Сессии"
// @Success 200 {file} file "report.csv"
// @Router /api/reports/{id}/export [get]
func (h *ReportHandler) ExportCSV(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	buf, err := h.service.ExportCSV(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s.csv", id.String()))
	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}
