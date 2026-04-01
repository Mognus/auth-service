package handler

import (
	client "auth-service/client"
	"auth-service/client/provider"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	provider *provider.UserProvider
	mw       *client.Middleware
}

func NewUserHandler(p *provider.UserProvider, mw *client.Middleware) *UserHandler {
	return &UserHandler{provider: p, mw: mw}
}

func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Use(h.mw.JWTMiddleware())
	users.Get("/", h.provider.ListHandler())
	users.Get("/schema", h.provider.SchemaHandler())
	users.Get("/:id", h.provider.GetHandler())
	users.Post("/", h.provider.CreateHandler())
	users.Put("/:id", h.provider.UpdateHandler())
	users.Delete("/:id", h.provider.DeleteHandler())
}
