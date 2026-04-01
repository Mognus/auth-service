package client

import "github.com/gofiber/fiber/v2"

type AuthService struct {
	handlers []Handler
}

func NewAuthService(handlers []Handler) *AuthService {
	return &AuthService{handlers: handlers}
}

func (s *AuthService) Name() string { return "auth" }

func (s *AuthService) RegisterRoutes(router fiber.Router) {
	for _, h := range s.handlers {
		h.RegisterRoutes(router)
	}
}
