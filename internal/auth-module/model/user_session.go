package model

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"not null;default:false"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}