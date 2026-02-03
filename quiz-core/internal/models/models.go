package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type User struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email            string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash     string    `gorm:"not null" json:"-"`
	Role             string    `gorm:"default:'teacher'" json:"role"`
	ResetToken       string    `json:"-"`
	ResetTokenExpiry time.Time `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
}

type Quiz struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TeacherID   uuid.UUID      `gorm:"type:uuid;not null" json:"teacher_id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Settings    datatypes.JSON `json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	Questions   []Question     `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE" json:"questions"`
}

type Question struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	QuizID   int64     `json:"quiz_id"`
	Type     string    `gorm:"not null" json:"type"`
	Text     string    `gorm:"not null" json:"text"`
	TimerSec int       `json:"timer_sec"`
	Score    int       `json:"score"`
	Order    int       `json:"order"`
	Options  []Option  `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"options"`
}

type Option struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	Text       string    `gorm:"not null" json:"text"`
	IsCorrect  bool      `json:"is_correct"`
}

type GameSession struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"session_id"`
	QuizID         int64          `json:"quiz_id"`
	HostID         uuid.UUID      `json:"host_id"`
	StartedAt      time.Time      `json:"started_at"`
	FinishedAt     time.Time      `json:"finished_at"`
	ReportSnapshot datatypes.JSON `json:"report_snapshot"`
}

type ReportSnapshot struct {
	Participants []struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
		Rank  int    `json:"rank"`
		Answers map[string]bool `json:"answers"` 
	} `json:"participants"`
	QuestionIDs []string `json:"question_ids"` 
}
