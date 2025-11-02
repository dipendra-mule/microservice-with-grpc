package user

import (
	"context"
	"errors"

	"github.com/dipendra-mule/microservice-with-grpc/pkg/auth"
	"github.com/dipendra-mule/microservice-with-grpc/proto/user"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo       *Repository
	jwtManager *auth.JWTManager
}

func NewService(repo *Repository, jwtManager *auth.JWTManager) *Service {
	return &Service{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *Service) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	return s.repo.CreateUser(ctx, req)
}

func (s *Service) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.User, error) {
	return s.repo.GetUserByID(ctx, req.Id)
}

func (s *Service) Authenticate(ctx context.Context, req *user.AuthRequest) (*user.AuthResponse, error) {
	u, passwordHash, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.jwtManager.Generate(u.Id, u.Email, u.Role)
	if err != nil {
		return nil, err
	}

	return &user.AuthResponse{
		Token: token,
		User:  u,
	}, nil
}

func (s *Service) ValidateToken(ctx context.Context, req *user.ValidateTokenRequest) (*user.ValidateTokenResponse, error) {
	claims, err := s.jwtManager.Verify(req.Token)
	if err != nil {
		return &user.ValidateTokenResponse{Valid: false}, nil
	}

	u, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return &user.ValidateTokenResponse{Valid: false}, nil
	}

	return &user.ValidateTokenResponse{
		Valid: true,
		User:  u,
	}, nil
}

func (s *Service) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	users, total, err := s.repo.ListUsers(ctx, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	return &user.ListUsersResponse{
		Users: users,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

// implement UpdateUser method
