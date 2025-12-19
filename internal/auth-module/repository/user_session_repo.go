package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"admin-portal/internal/auth-module/model"
)

type UserSessionRepository interface {
	Create(ctx context.Context, token *model.UserSession) error
	FindValid(ctx context.Context, token string) (*model.UserSession, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) UserSessionRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *model.UserSession) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindValid(ctx context.Context, token string) (*model.UserSession, error) {
	var rt model.UserSession
	err := r.db.WithContext(ctx).
		Where("token = ? AND is_revoked = FALSE AND expires_at > ?", token, time.Now()).
		First(&rt).Error
	return &rt, err
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error
}
