package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dipendra-mule/microservice-with-grpc/proto/order"
	"github.com/google/uuid"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateOrder(ctx context.Context, o *order.Order) (*order.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate order ID if not set
	if o.Id == "" {
		o.Id = uuid.New().String()
	}

	now := time.Now()

	// Insert order
	orderQuery := `
        INSERT INTO orders (id, user_id, total_amount, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, user_id, total_amount, status, created_at, updated_at
    `

	var createdOrder order.Order
	err = tx.QueryRowContext(ctx, orderQuery,
		o.Id, o.UserId, o.TotalAmount, o.Status, now, now,
	).Scan(
		&createdOrder.Id, &createdOrder.UserId, &createdOrder.TotalAmount,
		&createdOrder.Status, &createdOrder.CreatedAt, &createdOrder.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Insert order items
	itemQuery := `
        INSERT INTO order_items (id, order_id, product_id, quantity, price, product_name)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	for _, item := range o.Items {
		itemID := uuid.New().String()
		_, err = tx.ExecContext(ctx, itemQuery,
			itemID, o.Id, item.ProductId, item.Quantity, item.Price, item.ProductName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdOrder.Items = o.Items
	return &createdOrder, nil
}

func (r *Repository) GetOrderByID(ctx context.Context, id string) (*order.Order, error) {
	var o order.Order

	orderQuery := `
        SELECT id, user_id, total_amount, status, created_at, updated_at
        FROM orders
        WHERE id = $1
		`
	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&o.Id, &o.UserId, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	itemsQuery := `
		SELECT product_id, quantity, price, product_name
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	defer rows.Close()

	var items []*order.OrderItem
	for rows.Next() {
		var item order.OrderItem
		if err := rows.Scan(&item.ProductId, &item.Quantity, &item.Price, &item.ProductName); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, &item) // Append item to list
	}
	o.Items = items // Set items to order
	return &o, nil
}

func (r *Repository) ListOrders(ctx context.Context, userID string, pag, limit int32) ([]*order.Order, int32, error) {
	offset := (pag - 1) * limit

	listQuery := `
		SELECT id, user_id, total_amount, status, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, listQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	// Create orders list
	var orders []*order.Order
	for rows.Next() {
		var o order.Order
		if err := rows.Scan(&o.Id, &o.UserId, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &o)
	}

	// Get total count
	var total int32
	countQuery := `
		SELECT COUNT(*) FROM orders WHERE user_id = $1
		`
	err = r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count: %w", err)
	}

	return orders, total, nil
}

// UpdateOrderStatus updates the status of an order
// and returns the updated order and error
func (r *Repository) UpdateOrderStatus(ctx context.Context, id, status string) (*order.Order, error) {
	updateStatusQuery := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, user_id, total_amount, status, created_at, updated_at
	`
	var o order.Order
	err := r.db.QueryRowContext(ctx, updateStatusQuery, status, time.Now(), id).Scan(
		&o.Id, &o.UserId, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}
	return &o, nil
}
