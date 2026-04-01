package handler

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	authv1 "auth-service/gen/auth/v1"
	client "auth-service/client"
	apperrors "github.com/Mognus/go-grpc-crud/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type AuthHandler struct {
	grpcClient authv1.AuthServiceClient
	config     *client.Config
}

func NewAuthHandler(grpcClient authv1.AuthServiceClient, config *client.Config) *AuthHandler {
	return &AuthHandler{grpcClient: grpcClient, config: config}
}

func (h *AuthHandler) RegisterRoutes(router fiber.Router) {
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 15 * time.Minute,
		Storage:    h.config.Storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "login:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return apperrors.TooManyRequests("Too many login attempts. Try again in 15 minutes.")
		},
	})

	authRoutes := router.Group("/auth")
	authRoutes.Post("/login", loginLimiter, h.Login)
	authRoutes.Post("/register", h.Register)
	authRoutes.Post("/refresh", h.Refresh)
	authRoutes.Post("/logout", h.Logout)
	authRoutes.Get("/me", h.config.JWTMiddleware(), h.config.RequireAuth, h.Me)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("Invalid request body")
	}
	resp, err := h.grpcClient.Login(c.UserContext(), &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}
	setAccessCookie(c, resp.AccessToken)
	setRefreshCookie(c, resp.RefreshToken)
	return c.JSON(fiber.Map{"accessToken": resp.AccessToken, "user": client.ToUserJSON(resp.User)})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("Invalid request body")
	}
	resp, err := h.grpcClient.Register(c.UserContext(), &authv1.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}
	setAccessCookie(c, resp.AccessToken)
	setRefreshCookie(c, resp.RefreshToken)
	return c.Status(201).JSON(fiber.Map{"accessToken": resp.AccessToken, "user": client.ToUserJSON(resp.User)})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return apperrors.Unauthorized("No refresh token")
	}
	resp, err := h.grpcClient.RefreshToken(c.UserContext(), &authv1.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}
	setAccessCookie(c, resp.AccessToken)
	setRefreshCookie(c, resp.RefreshToken)
	return c.JSON(fiber.Map{"accessToken": resp.AccessToken, "refreshToken": resp.RefreshToken})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, err := client.GetUserIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.grpcClient.GetUser(c.UserContext(), &authv1.GetUserRequest{Id: uint64(userID)})
	if err != nil {
		return apperrors.GrpcToHTTP(err)
	}
	return c.JSON(client.ToUserJSON(resp.User))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		h.grpcClient.Logout(c.UserContext(), &authv1.LogoutRequest{RefreshToken: refreshToken})
	}
	clearCookie(c, "access_token")
	clearCookie(c, "refresh_token")
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

func setAccessCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  jwtExpiry(token, 15*time.Minute),
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func setRefreshCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  jwtExpiry(token, 7*24*time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func clearCookie(c *fiber.Ctx, name string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func jwtExpiry(token string, fallback time.Duration) time.Time {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Now().Add(fallback)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Now().Add(fallback)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return time.Now().Add(fallback)
	}
	return time.Unix(claims.Exp, 0)
}
