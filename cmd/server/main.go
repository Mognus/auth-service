package main

import (
	"log"
	"net"
	"os"

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

	if err := database.AutoMigrate(&model.Role{}, &model.User{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	for _, role := range model.DefaultRoles() {
		database.FirstOrCreate(&role, model.Role{Name: role.Name})
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	authv1.RegisterAuthServiceServer(s, handler.New(database, jwtSecret))

	log.Println("auth-service listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
