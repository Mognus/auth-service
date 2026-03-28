package main

import (
	"log"
	"net"
	"os"
	"time"

	authv1 "auth-service/gen/auth/v1"
	"auth-service/internal/db"
	"auth-service/internal/handler"
	"auth-service/internal/model"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load()

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.AutoMigrate(&model.Role{}, &model.User{}, &model.RefreshToken{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	for _, role := range model.DefaultRoles() {
		database.FirstOrCreate(&role, model.Role{Name: role.Name})
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	accessTTL, err := time.ParseDuration(os.Getenv("JWT_ACCESS_TTL"))
	if err != nil {
		accessTTL = 15 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(os.Getenv("JWT_REFRESH_TTL"))
	if err != nil {
		refreshTTL = 7 * 24 * time.Hour
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	authv1.RegisterAuthServiceServer(s, handler.New(database, jwtSecret, accessTTL, refreshTTL))

	log.Println("auth-service listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
