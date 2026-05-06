package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"auth-service/internal/domain/users"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func GenerateTokenPair(ctx context.Context, db *gorm.DB, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration, user *users.User) (accessToken, refreshToken string, err error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role.Name,
		"exp":     time.Now().Add(accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(jwtSecret))
	if err != nil {
		return
	}

	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return
	}
	refreshToken = hex.EncodeToString(raw)

	rt := RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}
	err = db.WithContext(ctx).Create(&rt).Error
	return
}
