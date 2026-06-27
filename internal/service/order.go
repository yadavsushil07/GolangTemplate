package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type OrderService struct {
	db          *pgxpool.Pool
	orderRepo   *repository.OrderRepository
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewOrderService(db *pgxpool.Pool, orderRepo *repository.OrderRepository, cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *OrderService {
	return &OrderService{db: db, orderRepo: orderRepo, cartRepo: cartRepo, productRepo: productRepo}
}

func (s *OrderService) Checkout(ctx context.Context, userID int64, req model.CheckoutRequest) (*model.Order, error) {
	if req.ShippingName == "" || req.ShippingAddress == "" {
		return nil, fmt.Errorf("shipping name and address are required")
	}

	items, err := s.cartRepo.GetItems(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	totalCents := 0
	for _, item := range items {
		if item.Product == nil {
			return nil, fmt.Errorf("product data missing in cart")
		}
		result, err := tx.Exec(ctx,
			`UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
			item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}
		if result.RowsAffected() == 0 {
			return nil, fmt.Errorf("insufficient stock for product: %s", item.Product.Name)
		}
		totalCents += item.Product.PriceCents * item.Quantity
	}

	var order model.Order
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, total_cents, shipping_name, shipping_address)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, total_cents, status, shipping_name, shipping_address, created_at`,
		userID, totalCents, req.ShippingName, req.ShippingAddress,
	).Scan(&order.ID, &order.UserID, &order.TotalCents, &order.Status, &order.ShippingName, &order.ShippingAddress, &order.CreatedAt)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, price_cents)
			VALUES ($1, $2, $3, $4)`,
			order.ID, item.ProductID, item.Quantity, item.Product.PriceCents)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	_ = s.cartRepo.Clear(ctx, req.SessionID)
	return &order, nil
}

func (s *OrderService) ListByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID)
}

func (s *OrderService) ListAll(ctx context.Context) ([]model.Order, error) {
	return s.orderRepo.ListAll(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	switch status {
	case model.OrderStatusPlaced, model.OrderStatusShipped, model.OrderStatusDelivered, model.OrderStatusCancelled:
	default:
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.orderRepo.UpdateStatus(ctx, orderID, status)
}
