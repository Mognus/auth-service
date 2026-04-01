package client

import "github.com/gofiber/fiber/v2"

type Config struct {
	*Middleware
	Storage fiber.Storage
}

func NewConfig(jwtSecret string, storage fiber.Storage) *Config {
	return &Config{
		Middleware: NewMiddleware(jwtSecret),
		Storage:    storage,
	}
}
