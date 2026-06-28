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

func scanOrder(row interface {
	Scan(...any) error
}) (*model.Order, error) {
	o := &model.Order{}
	return o, row.Scan(
		&o.ID, &o.UserID, &o.TotalCents, &o.DiscountCents, &o.Status,
		&o.PaymentMethod, &o.PaymentStatus, &o.RazorpayOrderID, &o.CouponCode,
		&o.ShippingName, &o.ShippingAddress, &o.CustomizationNote, &o.CreatedAt,
	)
}

const orderSelectCols = `id, user_id, total_cents, discount_cents, status,
	payment_method, payment_status, razorpay_order_id, coupon_code,
	shipping_name, shipping_address, customization_note, created_at`

func (r *OrderRepository) Create(ctx context.Context, o *model.Order) (*model.Order, error) {
	return scanOrder(r.db.QueryRow(ctx, `
		INSERT INTO orders (user_id, total_cents, discount_cents, payment_method, payment_status,
		                    razorpay_order_id, coupon_code, shipping_name, shipping_address, customization_note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+orderSelectCols,
		o.UserID, o.TotalCents, o.DiscountCents, o.PaymentMethod, o.PaymentStatus,
		o.RazorpayOrderID, o.CouponCode, o.ShippingName, o.ShippingAddress, o.CustomizationNote,
	))
}

func (r *OrderRepository) AddItem(ctx context.Context, orderID, productID int64, variantID *int64, qty, priceCents int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO order_items (order_id, product_id, variant_id, quantity, price_cents)
		VALUES ($1, $2, $3, $4, $5)`,
		orderID, productID, variantID, qty, priceCents)
	return err
}

func (r *OrderRepository) UpdateRazorpayOrderID(ctx context.Context, orderID int64, razorpayOrderID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET razorpay_order_id = $1 WHERE id = $2`,
		razorpayOrderID, orderID)
	return err
}

func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, orderID int64, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET payment_status = $1 WHERE id = $2`, status, orderID)
	return err
}

func (r *OrderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderSelectCols+` FROM orders WHERE id = $1`, id))
}

func (r *OrderRepository) FindByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (*model.Order, error) {
	return scanOrder(r.db.QueryRow(ctx,
		`SELECT `+orderSelectCols+` FROM orders WHERE razorpay_order_id = $1`, razorpayOrderID))
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+orderSelectCols+` FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *OrderRepository) ListAll(ctx context.Context) ([]model.Order, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+orderSelectCols+` FROM orders ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, orderID)
	return err
}

func scanOrders(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]model.Order, error) {
	var orders []model.Order
	for rows.Next() {
		o := model.Order{}
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.TotalCents, &o.DiscountCents, &o.Status,
			&o.PaymentMethod, &o.PaymentStatus, &o.RazorpayOrderID, &o.CouponCode,
			&o.ShippingName, &o.ShippingAddress, &o.CustomizationNote, &o.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
