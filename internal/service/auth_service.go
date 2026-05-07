package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/domain/auth"
	"auth-service/internal/domain/users"
	"auth-service/internal/repository"

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
	auths           *repository.AuthRepository
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
		auths:           repository.NewAuthRepository(db),
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.auths.FindUserByEmail(ctx, email)
	if err != nil {
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
	if _, err := s.auths.FindUserByEmail(ctx, email); err == nil {
		return nil, ErrUserAlreadyExists
	}

	defaultRole, err := s.auths.FindDefaultRole(ctx)
	if err != nil {
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
	if err := s.auths.CreateUser(ctx, &user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateUser, err)
	}
	user.Role = defaultRole

	return s.issueTokenPair(ctx, &user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	rt, err := s.auths.FindValidRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.auths.FindUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Refresh tokens are single-use; keep the existing best-effort revoke behavior.
	s.auths.RevokeRefreshTokenByID(ctx, rt.ID)

	accessToken, newRefreshToken, err := s.issueTokens(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerateTokens, err)
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.auths.RevokeRefreshToken(ctx, refreshToken)
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *users.User) (*AuthResult, error) {
	accessToken, refreshToken, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerateTokens, err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *users.User) (accessToken, refreshToken string, err error) {
	accessToken, err = auth.GenerateAccessToken(s.jwtSecret, s.accessTokenTTL, user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = auth.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	err = s.auths.CreateRefreshToken(ctx, &auth.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
