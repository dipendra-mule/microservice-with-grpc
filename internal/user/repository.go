package user

import (
	"context"
	"database/sql"
	"errors"

	"time"

	"github.com/dipendra-mule/microservice-with-grpc/proto/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	// Check if email already exists
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var u user.User
	query := `
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, name, role, created_at, updated_at
	`
	now := time.Now()
	err = r.db.QueryRowContext(ctx, query, req.Email, string(hashedPassword), req.Name, req.Role, now, now).Scan(&u.Id, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	query := `
		SELECT id, email, name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.Id, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*user.User, string, error) {
	var u user.User
	var passwordHash string

	query := `
		SELECT id, email, name, role, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.Id, &u.Email, &u.Name, &u.Role, &passwordHash, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, "", ErrUserNotFound
	}

	if err != nil {
		return nil, "", err
	}
	return &u, passwordHash, nil
}

func (r *Repository) ListUsers(ctx context.Context, page, limit int32) ([]*user.User, int32, error) {
	offset := (page - 1) * limit

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, name, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.Id, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	var total int32
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
