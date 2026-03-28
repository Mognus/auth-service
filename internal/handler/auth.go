package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	db             *gorm.DB
	jwtSecret      string
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
}

func New(db *gorm.DB, jwtSecret string, accessTTL, refreshTTL time.Duration) *Handler {
	return &Handler{db: db, jwtSecret: jwtSecret, accessTokenTTL: accessTTL, refreshTokenTTL: refreshTTL}
}

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	var user model.User
	if err := h.db.WithContext(ctx).Preload("Role").Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	if !user.Active {
		return nil, status.Error(codes.PermissionDenied, "account is deactivated")
	}
	if !user.CheckPassword(req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	accessToken, refreshToken, err := h.generateTokenPair(ctx, &user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	return &authv1.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: toUserResponse(&user)}, nil
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	var existing model.User
	if err := h.db.WithContext(ctx).Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}
	if len(req.Password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	var defaultRole model.Role
	if err := h.db.WithContext(ctx).Where("name = ?", string(model.RoleUser)).First(&defaultRole).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to find default role")
	}

	user := model.User{
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

	accessToken, refreshToken, err := h.generateTokenPair(ctx, &user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	return &authv1.RegisterResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: toUserResponse(&user)}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	var rt model.RefreshToken
	err := h.db.WithContext(ctx).
		Where("token = ? AND revoked = false AND expires_at > ?", req.RefreshToken, time.Now()).
		First(&rt).Error
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}

	var user model.User
	if err := h.db.WithContext(ctx).Preload("Role").First(&user, rt.UserID).Error; err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	// Rotation: revoke the old token and issue a new pair.
	// If someone steals the refresh token and uses it, the legitimate user's
	// token becomes invalid on their next refresh — a detectable signal of theft.
	h.db.WithContext(ctx).Model(&rt).Update("revoked", true)

	accessToken, newRefreshToken, err := h.generateTokenPair(ctx, &user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	// The new refresh token must be returned so the client can replace the old one.
	// Without this, the next refresh would fail since the old token is now revoked.
	return &authv1.RefreshTokenResponse{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

func (h *Handler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	result := h.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token = ?", req.RefreshToken).
		Update("revoked", true)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	return &authv1.LogoutResponse{Success: true}, nil
}

func (h *Handler) generateTokenPair(ctx context.Context, user *model.User) (accessToken, refreshToken string, err error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role.Name,
		"exp":     time.Now().Add(h.accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return
	}

	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return
	}
	refreshToken = hex.EncodeToString(raw)

	rt := model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(h.refreshTokenTTL),
	}
	err = h.db.WithContext(ctx).Create(&rt).Error
	return
}
