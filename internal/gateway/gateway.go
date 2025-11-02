package gateway

import (
	"context"
	"net/http"

	"github.com/dipendra-mule/microservice-with-grpc/internal/user"
	"github.com/dipendra-mule/microservice-with-grpc/proto/user"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	userServiceAddr    string
	orderServiceAddr   string
	productServiceAddr string
}

func NewGateway(userAddr, orderAddr, productAddr string) *Gateway {
	return &Gateway{
		userServiceAddr:    userAddr,
		orderServiceAddr:   orderAddr,
		productServiceAddr: productAddr,
	}
}

func (g *Gateway) Start(port string) error {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register services
	if err := user.RegisterUserServiceHandlerFromEndpoint(ctx, mux, g.userServiceAddr, opts); err != nil {
		return err
	}
	if err := order.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, g.orderServiceAddr, opts); err != nil {
		return err
	}
	if err := product.RegisterProductServiceHandlerFromEndpoint(ctx, mux, g.productServiceAddr, opts); err != nil {
		return err
	}

	// Add CORS middleware
	handler := corsMiddleware(mux)

	return http.ListenAndServe(":"+port, handler)
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
