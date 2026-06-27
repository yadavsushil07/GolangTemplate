package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, userID int64, totalCents int, shippingName, shippingAddress string) (*model.Order, error) {
	o := &model.Order{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO orders (user_id, total_cents, shipping_name, shipping_address)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, total_cents, status, shipping_name, shipping_address, created_at`,
		userID, totalCents, shippingName, shippingAddress,
	).Scan(&o.ID, &o.UserID, &o.TotalCents, &o.Status, &o.ShippingName, &o.ShippingAddress, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (r *OrderRepository) AddItem(ctx context.Context, orderID, productID int64, qty, priceCents int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO order_items (order_id, product_id, quantity, price_cents)
		VALUES ($1, $2, $3, $4)`,
		orderID, productID, qty, priceCents)
	return err
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, total_cents, status, shipping_name, shipping_address, created_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalCents, &o.Status, &o.ShippingName, &o.ShippingAddress, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepository) ListAll(ctx context.Context) ([]model.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, total_cents, status, shipping_name, shipping_address, created_at
		FROM orders ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalCents, &o.Status, &o.ShippingName, &o.ShippingAddress, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, orderID)
	return err
}
