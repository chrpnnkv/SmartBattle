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

func (r *ReportRepository) Save(session *models.GameSession) error {
	return r.db.Create(session).Error
}

func (r *ReportRepository) GetByHostID(hostID uuid.UUID) ([]models.GameSession, error) {
	var sessions []models.GameSession
	err := r.db.Where("host_id = ?", hostID).Order("started_at desc").Find(&sessions).Error
	return sessions, err
}

func (r *ReportRepository) GetByID(id uuid.UUID) (*models.GameSession, error) {
	var session models.GameSession
	err := r.db.First(&session, "id = ?", id).Error
	return &session, err
}
