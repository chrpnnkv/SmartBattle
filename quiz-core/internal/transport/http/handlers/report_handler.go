package handlers

import (
	"fmt"
	"net/http"

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
// @Success 200 {array} models.GameSession
// @Router /api/reports [get]
func (h *ReportHandler) GetReports(c *gin.Context) {
	hostID, _ := uuid.Parse(c.GetString("user_id"))
	reports, err := h.service.GetTeacherReports(hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
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
