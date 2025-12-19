package model

import (
	"time"

	"github.com/google/uuid"
)

type PasswordMaster struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`

	PasswordHash string `gorm:"type:text;not null"`

	IsActive bool `gorm:"not null;default:true"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`

	// Relations
	User User `gorm:"constraint:OnDelete:CASCADE;"`
}

func (PasswordMaster) TableName() string {
	return "password_master"
}
