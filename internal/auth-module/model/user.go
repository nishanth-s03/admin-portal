package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	Username string `gorm:"type:varchar(150);uniqueIndex;not null"`

	Role string `gorm:"type:varchar(20);not null;check:role IN ('user','admin','super-admin')"`

	IsActive    bool `gorm:"not null;default:true"`
	IsActivated bool `gorm:"not null;default:false"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`

	// Relations
	Passwords []PasswordMaster `gorm:"foreignKey:UserID"`
	LoginLogs []LoginLog       `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}
