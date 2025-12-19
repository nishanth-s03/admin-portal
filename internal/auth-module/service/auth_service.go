package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"admin-portal/internal/auth-module/model"
	"admin-portal/internal/auth-module/repository"
)

type AuthService interface {
	Register(ctx context.Context, username, password, role string) (*model.User, error)
	ActivateUser(ctx context.Context, userID string) error
	Login(ctx context.Context, username, password string) (*model.User, string, string, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	db           *gorm.DB
	userRepo     repository.UserRepository
	passwordRepo repository.PasswordRepository
	loginLogRepo repository.LoginLogRepository
	tokenService TokenService
}

func NewAuthService(
	db *gorm.DB,
	userRepo repository.UserRepository,
	passwordRepo repository.PasswordRepository,
	loginLogRepo repository.LoginLogRepository,
	tokenService TokenService,
) AuthService {
	return &authService{
		db:           db,
		userRepo:     userRepo,
		passwordRepo: passwordRepo,
		loginLogRepo: loginLogRepo,
		tokenService: tokenService,
	}
}

/* Register creates a new user with the given username, password, and role. */
func (s *authService) Register(
	ctx context.Context,
	username, password, role string) (*model.User, error) {
	//Check for user data existence
	_, err := s.userRepo.FindByUsername(ctx, username)

	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	//Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	//Create User
	var user *model.User

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user = &model.User{
			Username:    username,
			Role:        role,
			IsActive:    true,
			IsActivated: false,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}

		passwordEntry := &model.PasswordMaster{
			UserID:       user.ID,
			PasswordHash: string(hashedPassword),
			IsActive:     true,
		}

		if err := s.passwordRepo.Create(ctx, passwordEntry); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

/* ActivateUser sets the IsActivated flag of a user to true. */
func (s *authService) ActivateUser(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsActivated = true

	return s.userRepo.Update(ctx, user)
}

/* Login authenticates a user with the given username and password. */
func (s *authService) Login(
	ctx context.Context,
	username, password string,
) (*model.User, string, string, error) {

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		s.writeLoginLog(ctx, nil, "Invalid username", "error")
		return nil, "", "", ErrInvalidCredential
	}

	if !user.IsActive || !user.IsActivated {
		return nil, "", "", ErrUserInactive
	}

	pass, err := s.passwordRepo.FindActiveByUserID(ctx, user.ID.String())
	if err != nil {
		return nil, "", "", ErrInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(pass.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, "", "", ErrInvalidCredential
	}

	access, refresh, err := s.tokenService.IssueTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	s.writeLoginLog(ctx, &user.ID, "Login successful", "success")
	return user, access, refresh, nil
}

// Logout revokes the given refresh token.
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		// idempotent logout
		return nil
	}

	return s.tokenService.Logout(ctx, refreshToken)
}


/*------------------------------Helpers----------------------------------*/
func (s *authService) writeLoginLog(
	ctx context.Context,
	userID *uuid.UUID,
	message string,
	logType string,
) {
	_ = s.loginLogRepo.Create(ctx, &model.LoginLog{
		UserID: userID,
		Message: message,
		LogType: logType,
	})
}
