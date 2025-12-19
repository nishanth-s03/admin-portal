package model

import (
	"time"

	"github.com/google/uuid"
)

type LoginLog struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	UserID *uuid.UUID `gorm:"type:uuid;index"`

	Message string `gorm:"type:text;not null"`

	LogType string `gorm:"type:varchar(10);not null;check:log_type IN ('warn','error','info','success')"`

	IPAddress  *string `gorm:"type:inet"`
	UserAgent *string `gorm:"type:text"`

	IsDeleted bool `gorm:"not null;default:false"`

	CreatedAt time.Time `gorm:"not null;default:now()"`

	// Relations
	User *User `gorm:"constraint:OnDelete:SET NULL;"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}
