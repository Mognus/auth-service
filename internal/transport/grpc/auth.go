package grpc

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

	setTokenCookies(ctx, result.AccessToken, result.RefreshToken)

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

	setTokenCookies(ctx, result.AccessToken, result.RefreshToken)

	return &authv1.RegisterResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toUserResponse(result.User),
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	token := req.RefreshToken
	if token == "" {
		token = cookieFromMetadata(ctx, "refresh_token")
	}

	result, err := h.authService.RefreshToken(ctx, token)
	if err != nil {
		return nil, authErrorToStatus(err)
	}

	setTokenCookies(ctx, result.AccessToken, result.RefreshToken)

	return &authv1.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	token := req.RefreshToken
	if token == "" {
		token = cookieFromMetadata(ctx, "refresh_token")
	}

	if token != "" {
		h.authService.Logout(ctx, token)
	}

	clearCookies(ctx)

	return &authv1.LogoutResponse{Success: true}, nil
}

func setTokenCookies(ctx context.Context, accessToken, refreshToken string) {
	grpc.SetHeader(ctx, metadata.Pairs(
		"set-cookie", cookieString("access_token", accessToken, 15*time.Minute),
		"set-cookie", cookieString("refresh_token", refreshToken, 7*24*time.Hour),
	))
}

func clearCookies(ctx context.Context) {
	grpc.SetHeader(ctx, metadata.Pairs(
		"set-cookie", cookieString("access_token", "", -time.Hour),
		"set-cookie", cookieString("refresh_token", "", -time.Hour),
	))
}

func cookieString(name, value string, maxAge time.Duration) string {
	c := http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
	}
	return c.String()
}

func cookieFromMetadata(ctx context.Context, name string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, raw := range md.Get("cookie") {
		for _, part := range strings.Split(raw, ";") {
			part = strings.TrimSpace(part)
			if k, v, ok := strings.Cut(part, "="); ok && strings.TrimSpace(k) == name {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
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
