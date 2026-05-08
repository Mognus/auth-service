package token

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"auth-service/internal/domain/users"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateAccessToken(jwtSecret string, ttl time.Duration, user *users.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role.Name,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func GenerateRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
