package repository

import (
	"context"
	"time"

	"auth-service/internal/domain/auth"
	"auth-service/internal/domain/roles"
	"auth-service/internal/domain/users"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (users.User, error) {
	var user users.User
	err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *AuthRepository) FindUserByID(ctx context.Context, id uint) (users.User, error) {
	var user users.User
	err := r.db.WithContext(ctx).Preload("Role").First(&user, id).Error
	return user, err
}

func (r *AuthRepository) FindDefaultRole(ctx context.Context) (roles.Role, error) {
	var role roles.Role
	err := r.db.WithContext(ctx).Where("name = ?", string(roles.RoleUser)).First(&role).Error
	return role, err
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *AuthRepository) FindValidRefreshToken(ctx context.Context, token string) (auth.RefreshToken, error) {
	var refreshToken auth.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND revoked = false AND expires_at > ?", token, time.Now()).
		First(&refreshToken).Error
	return refreshToken, err
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&auth.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).
		Error
}

func (r *AuthRepository) RevokeRefreshTokenByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&auth.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked", true).
		Error
}
