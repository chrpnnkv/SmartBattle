package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type GameSession struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	QuizID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"quiz_id"`
	HostID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"host_id"`
	StartedAt      time.Time      `gorm:"not null" json:"started_at"`
	FinishedAt     time.Time      `json:"finished_at"`
	ReportSnapshot datatypes.JSON `gorm:"type:jsonb;not null" json:"report_snapshot"`
}
