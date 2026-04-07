package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email            string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash     string     `gorm:"type:varchar(255);not null" json:"-"`
	Role             string     `gorm:"type:varchar(50);default:'teacher'"`
	ResetToken       *string    `gorm:"type:varchar(255)" json:"-"`
	ResetTokenExpiry *time.Time `json:"-"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
