package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/auth"
	"auth-service/internal/roles"
	"auth-service/internal/users"

	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAccountDeactivated  = errors.New("account deactivated")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrDefaultRoleMissing  = errors.New("default role missing")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrCreateUser          = errors.New("create user")
	ErrGenerateTokens      = errors.New("generate tokens")
)

type AuthService struct {
	db              *gorm.DB
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         *users.User
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthService(db *gorm.DB, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		db:              db,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	var user users.User
	if err := s.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.Active {
		return nil, ErrAccountDeactivated
	}
	if !user.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, &user)
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (*AuthResult, error) {
	var existing users.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrUserAlreadyExists
	}

	var defaultRole roles.Role
	if err := s.db.WithContext(ctx).Where("name = ?", string(roles.RoleUser)).First(&defaultRole).Error; err != nil {
		return nil, ErrDefaultRoleMissing
	}

	user := users.User{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		RoleID:    defaultRole.ID,
		Active:    true,
	}
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateUser, err)
	}
	user.Role = defaultRole

	return s.issueTokenPair(ctx, &user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	var rt auth.RefreshToken
	err := s.db.WithContext(ctx).
		Where("token = ? AND revoked = false AND expires_at > ?", refreshToken, time.Now()).
		First(&rt).Error
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	var user users.User
	if err := s.db.WithContext(ctx).Preload("Role").First(&user, rt.UserID).Error; err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Refresh tokens are single-use; keep the existing best-effort revoke behavior.
	s.db.WithContext(ctx).Model(&rt).Update("revoked", true)

	accessToken, newRefreshToken, err := auth.GenerateTokenPair(
		ctx,
		s.db,
		s.jwtSecret,
		s.accessTokenTTL,
		s.refreshTokenTTL,
		&user,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerateTokens, err)
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	result := s.db.WithContext(ctx).
		Model(&auth.RefreshToken{}).
		Where("token = ?", refreshToken).
		Update("revoked", true)
	return result.Error
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *users.User) (*AuthResult, error) {
	accessToken, refreshToken, err := auth.GenerateTokenPair(
		ctx,
		s.db,
		s.jwtSecret,
		s.accessTokenTTL,
		s.refreshTokenTTL,
		user,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerateTokens, err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}
