package grpc

import (
	"context"
	"errors"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	result, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, authErrorToStatus(err)
	}

	return &authv1.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toUserResponse(result.User),
	}, nil
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := h.validate(req); err != nil {
		return nil, err
	}

	result, err := h.authService.Register(ctx, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return nil, authErrorToStatus(err)
	}

	return &authv1.RegisterResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toUserResponse(result.User),
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	result, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, authErrorToStatus(err)
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.authService.Logout(ctx, req.RefreshToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authv1.LogoutResponse{Success: true}, nil
}

func authErrorToStatus(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid email or password")
	case errors.Is(err, service.ErrAccountDeactivated):
		return status.Error(codes.PermissionDenied, "account is deactivated")
	case errors.Is(err, service.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user with this email already exists")
	case errors.Is(err, service.ErrDefaultRoleMissing):
		return status.Error(codes.Internal, "failed to find default role")
	case errors.Is(err, service.ErrCreateUser):
		return status.Error(codes.Internal, "failed to create user")
	case errors.Is(err, service.ErrInvalidRefreshToken):
		return status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	case errors.Is(err, service.ErrGenerateTokens):
		return status.Error(codes.Internal, "failed to generate tokens")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
