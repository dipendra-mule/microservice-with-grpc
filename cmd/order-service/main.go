package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/dipendra-mule/microservice-with-grpc/internal/order"
	"github.com/dipendra-mule/microservice-with-grpc/pkg/database"
	orderv1 "github.com/dipendra-mule/microservice-with-grpc/proto/order"
	productv1 "github.com/dipendra-mule/microservice-with-grpc/proto/product"
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
		DBName:   getEnv("DB_NAME", "microservices"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repository
	orderRepo := order.NewRepository(db)

	// Create gRPC connections to other services
	userConn, err := grpc.Dial(
		getEnv("USER_SERVICE_ADDR", "localhost:50051"),
		grpc.WithInsecure(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer userConn.Close()

	productConn, err := grpc.Dial(
		getEnv("PRODUCT_SERVICE_ADDR", "localhost:50053"),
		grpc.WithInsecure(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to product service: %v", err)
	}
	defer productConn.Close()

	// Create service clients
	userClient := userv1.NewUserServiceClient(userConn)
	productClient := productv1.NewProductServiceClient(productConn)

	// Initialize service and server
	orderService := order.NewService(orderRepo, productClient, userClient)
	orderServer := order.NewServer(orderService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, orderServer)

	// Start server
	port := getEnv("ORDER_SERVICE_PORT", "50052")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Order service starting on port %s", port)
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
