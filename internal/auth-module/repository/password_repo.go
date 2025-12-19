package repository

import (
	"context"

	"gorm.io/gorm"

	"admin-portal/internal/auth-module/model"
)

type PasswordRepository interface {
	Create(ctx context.Context, password *model.PasswordMaster) error
	DeactivateAllForUser(ctx context.Context, userID string) error
	FindActiveByUserID(ctx context.Context, userID string) (*model.PasswordMaster, error)
}

type passwordRepository struct {
	db *gorm.DB
}

func NewPasswordRepository(db *gorm.DB) PasswordRepository {
	return &passwordRepository{db: db}
}

func (r *passwordRepository) Create(ctx context.Context, password *model.PasswordMaster) error {
	return r.db.WithContext(ctx).Create(password).Error
}

func (r *passwordRepository) DeactivateAllForUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.PasswordMaster{}).
		Where("user_id = ? AND is_active = TRUE", userID).
		Update("is_active", false).Error
}

func (r *passwordRepository) FindActiveByUserID(ctx context.Context, userID string) (*model.PasswordMaster, error) {
	var password model.PasswordMaster
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = TRUE", userID).
		First(&password).Error

	if err != nil {
		return nil, err
	}
	return &password, nil
}
