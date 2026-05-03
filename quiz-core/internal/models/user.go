package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name             string     `gorm:"type:varchar(255)" json:"name"`
	Email            string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash     string     `gorm:"type:varchar(255);not null" json:"-"`
	Role             string     `gorm:"type:varchar(50);default:teacher" json:"role"`
	ResetToken       *string    `gorm:"type:varchar(255)" json:"-"`
	ResetTokenExpiry *time.Time `json:"-"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Role == "" {
		u.Role = "teacher"
	}
	return
}
