package repository

import (
	"context"

	"gorm.io/gorm"

	"admin-portal/internal/auth-module/model"
)

type LoginLogRepository interface {
	Create(ctx context.Context, log *model.LoginLog) error
	GetAllByUserID(ctx context.Context, userID string) ([]*model.LoginLog, error)
	GetAll(ctx context.Context) ([]*model.LoginLog, error)
}

type loginLogRepository struct {
	db *gorm.DB
}

func NewLoginLogRepository(db *gorm.DB) LoginLogRepository {
	return &loginLogRepository{db: db}
}

func (r *loginLogRepository) Create(ctx context.Context, log *model.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *loginLogRepository) GetAllByUserID(ctx context.Context, userID string) ([]*model.LoginLog, error) {
	var logs []*model.LoginLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *loginLogRepository) GetAll(ctx context.Context) ([]*model.LoginLog, error) {
	var logs []*model.LoginLog
	err := r.db.WithContext(ctx).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

