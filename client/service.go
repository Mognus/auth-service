package client

import (
	libcrud "github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

type Handler interface {
	RegisterRoutes(router fiber.Router)
}

type AuthService struct {
	Config    *Config
	Providers []libcrud.GRPCProvider
	handlers  []Handler
	conn      *grpc.ClientConn
}

func New(addr, jwtSecret string, storage fiber.Storage) (*AuthService, error) {
	grpcClient, conn, err := NewClient(addr)
	if err != nil {
		return nil, err
	}
	config := NewConfig(jwtSecret, storage)
	users := NewUserProvider(grpcClient)
	roles := NewRoleProvider(grpcClient)
	return &AuthService{
		Config:    config,
		Providers: []libcrud.GRPCProvider{users, roles},
		handlers:  []Handler{newAuthHandler(grpcClient, config)},
		conn:      conn,
	}, nil
}

func (s *AuthService) Close() { s.conn.Close() }

func (s *AuthService) Name() string { return "auth" }

func (s *AuthService) RegisterRoutes(router fiber.Router) {
	for _, h := range s.handlers {
		h.RegisterRoutes(router)
	}
}
