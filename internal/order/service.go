package order

import (
	"context"
	"fmt"

	"github.com/dipendra-mule/microservice-with-grpc/proto/order"
	"github.com/dipendra-mule/microservice-with-grpc/proto/product"
	"github.com/dipendra-mule/microservice-with-grpc/proto/user"
)

type Service struct {
	repo          *Repository
	productClient product.ProductServiceClient
	userClient    user.UserServiceClient
}

func NewService(r *Repository, pc product.ProductServiceClient, uc user.UserServiceClient) *Service {
	return &Service{
		repo:          r,
		productClient: pc,
		userClient:    uc,
	}
}

func (s *Service) CreateOrder(ctx context.Context, r *order.CreateOrderRequest) (*order.Order, error) {
	// Validate products and get product details
	productReqs := make([]*product.ProductValidation, len(r.Items))
	for i, item := range r.Items {
		productReqs[i] = &product.ProductValidation{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	validationReq, err := s.productClient.ValidateProducts(ctx, &product.ValidateProductsRequest{
		Items: productReqs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to validate products: %w", err)
	}

	if !validationReq.Valid {
		return nil, fmt.Errorf("invalid products in order")
	}

	// Calculate total and prepare order items
	var total float32
	orderItems := make([]*order.OrderItem, len(r.Items))
	productMap := make(map[string]*product.Product)
	for _, p := range validationReq.Products {
		productMap[p.Id] = p
	}

	for i, item := range r.Items {
		p := productMap[item.ProductId]
		orderItems[i] = &order.OrderItem{
			ProductId:   item.ProductId,
			Quantity:    item.Quantity,
			Price:       p.Price,
			ProductName: p.Name,
		}
		total += p.Price * float32(item.Quantity)
	}

	// create order
	o := &order.Order{
		UserId:      r.UserId,
		Items:       orderItems,
		TotalAmount: total,
		Status:      "pending",
	}
	return s.repo.CreateOrder(ctx, o)
}

// GetOrder returns an order by ID
// and also returns the user associated with the order

func (s *Service) GetOrder(ctx context.Context, r *order.GetOrderRequest) (*order.Order, error) {
	o, err := s.repo.GetOrderByID(ctx, r.Id)
	if err != nil {
		return nil, err
	}

	// Get user information
	userResp, err := s.userClient.GetUser(ctx, &user.GetUserRequest{Id: o.UserId})
	if err != nil && userResp != nil {
		o.User = userResp.User
	}

	return o, nil
}

func (s *Service) ListOrders(ctx context.Context, r *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	orders, total, err := s.repo.ListOrders(ctx, r.UserId, r.Page, r.Limit)
	if err != nil {
		return nil, err
	}

	return &order.ListOrdersResponse{
		Orders: orders,
		Total:  total,
		Page:   r.Page,
		Limit:  r.Limit,
	}, nil
}

func (s *Service) UpdateOrderStatus(ctx context.Context, r *order.UpdateOrderStatusRequest) (*order.Order, error) {
	o, err := s.repo.GetOrderByID(ctx, r.OrderId)
	if err != nil {
		return nil, err
	}

	// Update order status
	o.Status = r.Status
	updatedOrder, err := s.repo.UpdateOrderStatus(ctx, o.Id, o.Status)
	if err != nil {
		return nil, err
	}
	return updatedOrder, nil
}
