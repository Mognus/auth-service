package main

import (
	"context"
	"log"
	"log/slog"
	"net"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/config"
	"auth-service/internal/platform/db"
	grpcserver "auth-service/internal/transport/grpc"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	lis, err := net.Listen("tcp", cfg.Server.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(recoveryInterceptor, loggingInterceptor))
	authv1.RegisterAuthServiceServer(s, grpcserver.New(database, cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL))

	log.Printf("auth-service listening on %s", cfg.Server.Addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func recoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error(info.FullMethod+" panic", "panic", r)
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil && status.Code(err) == codes.Internal {
		slog.Error(info.FullMethod, "error", err)
	}
	return resp, err
}
