package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Quiz struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TeacherID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"authorId"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"type:varchar(50);default:draft" json:"status"`
	Mode        string         `gorm:"type:varchar(50);default:teacher_paced" json:"mode"`
	Settings    datatypes.JSON `gorm:"type:jsonb" json:"settings"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	Questions   []Question     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"questions"`
}

func (q *Quiz) BeforeCreate(tx *gorm.DB) (err error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return
}

type Question struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	QuizID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"quizId"`
	Type               string         `gorm:"type:varchar(50);not null" json:"type"`
	Text               string         `gorm:"type:text;not null" json:"text"`
	ImageURL           string         `gorm:"type:text" json:"imageUrl"`
	TimerSec           int            `gorm:"not null" json:"timeLimitSeconds"`
	Score              int            `gorm:"not null" json:"score"`
	Order              int            `gorm:"not null" json:"order"`
	Options            []Option       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"options"`
	CorrectTextAnswers datatypes.JSON `gorm:"type:jsonb" json:"correctTextAnswers,omitempty"`
}

func (q *Question) BeforeCreate(tx *gorm.DB) (err error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return
}

type Option struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index" json:"questionId"`
	Text       string    `gorm:"type:text;not null" json:"text"`
	IsCorrect  bool      `gorm:"not null" json:"isCorrect"`
	Color      string    `gorm:"type:varchar(50)" json:"color"`
}

func (o *Option) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}
