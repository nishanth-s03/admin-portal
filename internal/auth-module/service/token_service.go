package service

import (
	"context"

	"admin-portal/internal/auth-module/model"
	"admin-portal/internal/auth-module/repository"
	"admin-portal/internal/shared/security"
)

type TokenService interface {
	IssueTokens(ctx context.Context, user *model.User) (access, refresh string, err error)
	Logout(ctx context.Context, refreshToken string) error
}

type tokenService struct {
	cfg        security.JWTConfig
	refreshRepo repository.UserSessionRepository
}

func NewTokenService(
	cfg security.JWTConfig,
	refreshRepo repository.UserSessionRepository,
) TokenService {
	return &tokenService{cfg: cfg, refreshRepo: refreshRepo}
}

func (s *tokenService) IssueTokens(
	ctx context.Context,
	user *model.User,
) (string, string, error) {

	access, err := security.GenerateAccessToken(
		s.cfg,
		user.ID.String(),
		user.Username,
		user.Role,
	)
	if err != nil {
		return "", "", err
	}

	refresh, expires, err := security.GenerateRefreshToken(s.cfg)
	if err != nil {
		return "", "", err
	}

	err = s.refreshRepo.Create(ctx, &model.UserSession{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: expires,
		IsRevoked: false,
	})
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *tokenService) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshRepo.Revoke(ctx, refreshToken)
}
