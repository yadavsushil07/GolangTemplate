package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type CouponRepository struct {
	db *pgxpool.Pool
}

func NewCouponRepository(db *pgxpool.Pool) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) FindByCode(ctx context.Context, code string) (*model.Coupon, error) {
	c := &model.Coupon{}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, discount_pct, discount_cents, min_order_cents, is_active, expires_at, created_at
		FROM coupons WHERE code = $1`, code,
	).Scan(&c.ID, &c.Code, &c.DiscountPct, &c.DiscountCents, &c.MinOrderCents, &c.IsActive, &c.ExpiresAt, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *CouponRepository) Create(ctx context.Context, req model.CreateCouponRequest) (*model.Coupon, error) {
	c := &model.Coupon{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO coupons (code, discount_pct, discount_cents, min_order_cents, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, discount_pct, discount_cents, min_order_cents, is_active, expires_at, created_at`,
		req.Code, req.DiscountPct, req.DiscountCents, req.MinOrderCents, req.ExpiresAt,
	).Scan(&c.ID, &c.Code, &c.DiscountPct, &c.DiscountCents, &c.MinOrderCents, &c.IsActive, &c.ExpiresAt, &c.CreatedAt)
	return c, err
}

func (r *CouponRepository) Deactivate(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE coupons SET is_active = FALSE WHERE id = $1`, id)
	return err
}

func (r *CouponRepository) List(ctx context.Context) ([]model.Coupon, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, discount_pct, discount_cents, min_order_cents, is_active, expires_at, created_at
		FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []model.Coupon
	for rows.Next() {
		var c model.Coupon
		if err := rows.Scan(&c.ID, &c.Code, &c.DiscountPct, &c.DiscountCents, &c.MinOrderCents, &c.IsActive, &c.ExpiresAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}
	return coupons, rows.Err()
}
