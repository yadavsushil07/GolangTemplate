package model

import "time"

type Coupon struct {
	ID             int64      `json:"id"`
	Code           string     `json:"code"`
	DiscountPct    *int       `json:"discount_pct"`
	DiscountCents  *int       `json:"discount_cents"`
	MinOrderCents  int        `json:"min_order_cents"`
	IsActive       bool       `json:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateCouponRequest struct {
	Code          string     `json:"code"`
	DiscountPct   *int       `json:"discount_pct"`
	DiscountCents *int       `json:"discount_cents"`
	MinOrderCents int        `json:"min_order_cents"`
	ExpiresAt     *time.Time `json:"expires_at"`
}

type ValidateCouponRequest struct {
	Code           string `json:"code"`
	OrderTotalCents int   `json:"order_total_cents"`
}

type ValidateCouponResponse struct {
	Valid          bool   `json:"valid"`
	DiscountCents  int    `json:"discount_cents"`
	Message        string `json:"message"`
}
