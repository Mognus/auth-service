package server

import (
	"context"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/auth"
	"auth-service/internal/roles"
	"auth-service/internal/users"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	var user users.User
	if err := h.db.WithContext(ctx).Preload("Role").Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	if !user.Active {
		return nil, status.Error(codes.PermissionDenied, "account is deactivated")
	}
	if !user.CheckPassword(req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(
		ctx,
		h.db,
		h.jwtSecret,
		h.accessTokenTTL,
		h.refreshTokenTTL,
		&user,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(&user),
	}, nil
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	var existing users.User
	if err := h.db.WithContext(ctx).Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}

	var defaultRole roles.Role
	if err := h.db.WithContext(ctx).Where("name = ?", string(roles.RoleUser)).First(&defaultRole).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to find default role")
	}

	user := users.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleID:    defaultRole.ID,
		Active:    true,
	}
	if err := h.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}
	user.Role = defaultRole

	accessToken, refreshToken, err := auth.GenerateTokenPair(
		ctx,
		h.db,
		h.jwtSecret,
		h.accessTokenTTL,
		h.refreshTokenTTL,
		&user,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &authv1.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(&user),
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	var rt auth.RefreshToken
	err := h.db.WithContext(ctx).
		Where("token = ? AND revoked = false AND expires_at > ?", req.RefreshToken, time.Now()).
		First(&rt).Error
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}

	var user users.User
	if err := h.db.WithContext(ctx).Preload("Role").First(&user, rt.UserID).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}

	h.db.WithContext(ctx).Model(&rt).Update("revoked", true)

	accessToken, newRefreshToken, err := auth.GenerateTokenPair(
		ctx,
		h.db,
		h.jwtSecret,
		h.accessTokenTTL,
		h.refreshTokenTTL,
		&user,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	result := h.db.WithContext(ctx).
		Model(&auth.RefreshToken{}).
		Where("token = ?", req.RefreshToken).
		Update("revoked", true)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}

	return &authv1.LogoutResponse{Success: true}, nil
}
