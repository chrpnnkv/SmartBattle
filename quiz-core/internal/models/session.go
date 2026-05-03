package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GameSession struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	QuizID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"quizId"`
	HostID         uuid.UUID      `gorm:"type:uuid;index" json:"hostId"`
	PIN            string         `gorm:"type:varchar(10);uniqueIndex" json:"pin"`
	Status         string         `gorm:"type:varchar(50);default:'waiting'" json:"status"`
	Mode           string         `gorm:"type:varchar(50);default:'teacher_paced'" json:"mode"`
	StartedAt      time.Time      `gorm:"autoCreateTime" json:"startedAt"`
	FinishedAt     *time.Time     `json:"finishedAt"`
	ReportSnapshot datatypes.JSON `gorm:"type:jsonb" json:"reportSnapshot"`
}

func (s *GameSession) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = "waiting"
	}
	if s.Mode == "" {
		s.Mode = "teacher_paced"
	}
	return
}
