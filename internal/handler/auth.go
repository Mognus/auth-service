package handler

import (
	"context"
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
	db        *gorm.DB
	jwtSecret string
}

func New(db *gorm.DB, jwtSecret string) *Handler {
	return &Handler{db: db, jwtSecret: jwtSecret}
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
	token, err := h.generateToken(&user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	return &authv1.LoginResponse{Token: token, User: toUserResponse(&user)}, nil
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

	token, err := h.generateToken(&user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	return &authv1.RegisterResponse{Token: token, User: toUserResponse(&user)}, nil
}

func (h *Handler) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role.Name,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
