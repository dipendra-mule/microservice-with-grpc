package order

import (
	"context"

	"github.com/dipendra-mule/microservice-with-grpc/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	order.UnimplementedOrderServiceServer
	service *Service
}

func NewServer(s *Service) *Server {
	return &Server{service: s}
}

func (s *Server) CreateOrder(ctx context.Context, r *order.CreateOrderRequest) (*order.OrderResponse, error) {
	createdOrder, err := s.service.CreateOrder(ctx, r)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &order.OrderResponse{Order: createdOrder}, nil
}

func (s *Server) GetOrder(ctx context.Context, r *order.GetOrderRequest) (*order.OrderResponse, error) {
	fetchedOrder, err := s.service.GetOrder(ctx, r)
	if err != nil {
		if err == ErrOrderNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &order.OrderResponse{Order: fetchedOrder}, nil
}

func (s *Server) ListOrders(ctx context.Context, r *order.UpdateOrderStatusRequest) (*order.OrderResponse, error) {
	updatedOrderStatus, err := s.service.UpdateOrderStatus(ctx, r)
	if err != nil {
		if err == ErrOrderNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &order.OrderResponse{Order: updatedOrderStatus}, nil
}
