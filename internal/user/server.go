package user

import (
	"context"

	"github.com/dipendra-mule/microservice-with-grpc/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	user.UnimplementedUserServiceServer
	service *Service
}

func NewServer(service *Service) *Server {
	return &Server{service: service}
}

func (s *Server) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserResponse, error) {
	u, err := s.service.CreateUser(ctx, req)
	if err != nil {
		if err == ErrEmailExists {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.UserResponse{User: u}, nil
}

func (s *Server) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
	u, err := s.service.GetUser(ctx, req)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.UserResponse{User: u}, nil
}

func (s *Server) Authenticate(ctx context.Context, req *user.AuthRequest) (*user.AuthResponse, error) {
	resp, err := s.service.Authenticate(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return resp, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *user.ValidateTokenRequest) (*user.ValidateTokenResponse, error) {
	return s.service.ValidateToken(ctx, req)
}

func (s *Server) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	return s.service.ListUsers(ctx, req)
}
