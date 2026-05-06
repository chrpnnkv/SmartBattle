package repository

import (
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) SaveSessionReport(session *models.GameSession) error {
	return r.db.Save(session).Error
}

func (r *ReportRepository) GetTeacherReports(hostID uuid.UUID) ([]models.GameSession, error) {
	var sessions []models.GameSession
	err := r.db.Where("host_id = ? AND status = ?", hostID, "finished").Order("started_at desc").Find(&sessions).Error
	return sessions, err
}

// GetAllReports — все завершённые сессии всех преподавателей.
// Используется обработчиком GetReports, когда вызывающий — администратор.
func (r *ReportRepository) GetAllReports() ([]models.GameSession, error) {
	var sessions []models.GameSession
	err := r.db.Where("status = ?", "finished").Order("started_at desc").Find(&sessions).Error
	return sessions, err
}

func (r *ReportRepository) GetByID(id uuid.UUID) (*models.GameSession, error) {
	var session models.GameSession
	err := r.db.First(&session, "id = ?", id).Error
	return &session, err
}

func (r *ReportRepository) GetByPIN(pin string) (*models.GameSession, error) {
	var session models.GameSession
	err := r.db.First(&session, "pin = ?", pin).Error
	return &session, err
}
