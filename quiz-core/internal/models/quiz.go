package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Quiz struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TeacherID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"teacher_id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Settings    datatypes.JSON `gorm:"type:jsonb" json:"settings"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	Questions   []Question     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"questions"`
}

type Question struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	QuizID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"-"`
	Type               string         `gorm:"type:varchar(50);not null" json:"type"`
	Text               string         `gorm:"type:text;not null" json:"text"`
	TimerSec           int            `gorm:"not null" json:"timer_sec"`
	Score              int            `gorm:"not null" json:"score"`
	Order              int            `gorm:"not null" json:"-"`
	Options            []Option       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"options"`
	CorrectTextAnswers datatypes.JSON `gorm:"type:jsonb" json:"correct_text_answers,omitempty"`
}

type Option struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Text       string    `gorm:"type:text;not null" json:"text"`
	IsCorrect  bool      `gorm:"not null" json:"is_correct"`
}
