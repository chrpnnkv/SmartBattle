package repository

import (
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *models.GameSession) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) GetByID(id uuid.UUID) (*models.GameSession, error) {
	var session models.GameSession
	err := r.db.First(&session, "id = ?", id).Error
	return &session, err
}

func (r *SessionRepository) GetByPIN(pin string) (*models.GameSession, error) {
	var session models.GameSession
	err := r.db.First(&session, "pin = ?", pin).Error
	return &session, err
}

func (r *SessionRepository) Update(session *models.GameSession) error {
	return r.db.Save(session).Error
}
