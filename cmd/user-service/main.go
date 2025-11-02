package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/dipendra-mule/microservice-with-grpc/internal/user"
	"github.com/dipendra-mule/microservice-with-grpc/pkg/auth"
	"github.com/dipendra-mule/microservice-with-grpc/pkg/database"
	userv1 "github.com/dipendra-mule/microservice-with-grpc/proto/user"
	"google.golang.org/grpc"
)

func main() {
	// Database configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "userservice"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repository and service
	userRepo := user.NewRepository(db)
	jwtManager := auth.NewJWTManager(
		getEnv("JWT_SECRET", "secret"),
		24*time.Hour, // 24 hours
	)

	userService := user.NewService(userRepo, jwtManager)
	userServer := user.NewServer(userService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, userServer)

	// Start server
	port := getEnv("USER_SERVICE_PORT", "50051")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("User service starting on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
