package client

import (
	apperrors "github.com/Mognus/go-grpc-crud/errors"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
	RoleGuest UserRole = "guest"
)

type Middleware struct {
	jwtMiddleware fiber.Handler
}

func NewMiddleware(jwtSecret string) *Middleware {
	m := &Middleware{}
	m.jwtMiddleware = jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(jwtSecret)},
		TokenLookup: "header:Authorization,cookie:access_token",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return apperrors.Unauthorized(err.Error())
		},
	})
	return m
}

func (m *Middleware) JWTMiddleware() fiber.Handler { return m.jwtMiddleware }

func (m *Middleware) RequireAuth(c *fiber.Ctx) error {
	if c.Locals("user") == nil {
		return apperrors.Unauthorized("Authentication required")
	}
	return c.Next()
}

func (m *Middleware) RequireAdmin(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return apperrors.Unauthorized("Authentication required")
	}
	token := user.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	if claims["role"].(string) != string(RoleAdmin) {
		return apperrors.Forbidden("Admin access required")
	}
	return c.Next()
}

func GetUserIDFromContext(c *fiber.Ctx) (uint, error) {
	user := c.Locals("user")
	if user == nil {
		return 0, apperrors.Unauthorized("Authentication required")
	}
	token := user.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	return uint(claims["user_id"].(float64)), nil
}

func GetUserRoleFromContext(c *fiber.Ctx) (UserRole, error) {
	user := c.Locals("user")
	if user == nil {
		return "", apperrors.Unauthorized("Authentication required")
	}
	token := user.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	return UserRole(claims["role"].(string)), nil
}
