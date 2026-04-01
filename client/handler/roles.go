package handler

import (
	client "auth-service/client"
	"auth-service/client/provider"

	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct {
	provider *provider.RoleProvider
	mw       *client.Middleware
}

func NewRoleHandler(p *provider.RoleProvider, mw *client.Middleware) *RoleHandler {
	return &RoleHandler{provider: p, mw: mw}
}

func (h *RoleHandler) RegisterRoutes(router fiber.Router) {
	roles := router.Group("/roles")
	roles.Use(h.mw.JWTMiddleware())
	roles.Get("/", h.provider.ListHandler())
	roles.Get("/schema", h.provider.SchemaHandler())
	roles.Get("/:id", h.provider.GetHandler())
	roles.Post("/", h.provider.CreateHandler())
	roles.Put("/:id", h.provider.UpdateHandler())
	roles.Delete("/:id", h.provider.DeleteHandler())
}
