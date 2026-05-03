package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chrpnnkv/SmartBattle/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportHandler struct {
	service     *service.ReportService
	quizService *service.QuizService
}

func NewReportHandler(svc *service.ReportService, quizSvc *service.QuizService) *ReportHandler {
	return &ReportHandler{service: svc, quizService: quizSvc}
}

// @Summary INTERNAL: Сохранить результаты игры
// @Tags Internal
// @Security InternalSecretAuth
// @Accept json
// @Produce json
// @Param request body service.RealtimeResultsPayload true "Итоги сессии из realtime"
// @Success 201
// @Router /internal/reports [post]
func (h *ReportHandler) SaveInternalReport(c *gin.Context) {
	var payload service.RealtimeResultsPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SaveRealtimeResults(&payload); err != nil {
		log.Printf("SaveInternalReport FAILED: room=%s quiz=%s err=%v",
			payload.RoomCode, payload.QuizID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("SaveInternalReport OK: room=%s quiz=%s participants=%d",
		payload.RoomCode, payload.QuizID, len(payload.Results))
	c.Status(http.StatusCreated)
}

// @Summary История игр преподавателя
// @Tags Analytics
// @Security BearerAuth
// @Produce json
// @Success 200 {array} GameReportDTO
// @Router /api/reports [get]
func (h *ReportHandler) GetReports(c *gin.Context) {
	hostIDStr := c.GetString("user_id")
	hostID, err := uuid.Parse(hostIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id in token"})
		return
	}

	sessions, err := h.service.GetTeacherReports(hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем для отладки: преподаватель → сколько отчётов нашли в БД.
	log.Printf("GetReports: host_id=%s found=%d", hostID, len(sessions))

	reports := make([]GameReportDTO, 0, len(sessions))
	for _, s := range sessions {
		reports = append(reports, h.mapToDTO(s))
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

	c.JSON(http.StatusOK, h.mapToDTO(*session))
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
