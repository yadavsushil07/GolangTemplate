package model

import "time"

const (
	OrderStatusPlaced    = "placed"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"

	PaymentMethodCOD      = "cod"
	PaymentMethodRazorpay = "razorpay"

	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
)

type Order struct {
	ID                 int64       `json:"id"`
	UserID             int64       `json:"user_id"`
	TotalCents         int         `json:"total_cents"`
	DiscountCents      int         `json:"discount_cents"`
	Status             string      `json:"status"`
	PaymentMethod      string      `json:"payment_method"`
	PaymentStatus      string      `json:"payment_status"`
	RazorpayOrderID    string      `json:"razorpay_order_id,omitempty"`
	CouponCode         string      `json:"coupon_code,omitempty"`
	ShippingName       string      `json:"shipping_name"`
	ShippingAddress    string      `json:"shipping_address"`
	CustomizationNote  string      `json:"customization_note,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	Items              []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID         int64           `json:"id"`
	OrderID    int64           `json:"order_id"`
	ProductID  int64           `json:"product_id"`
	VariantID  *int64          `json:"variant_id,omitempty"`
	Quantity   int             `json:"quantity"`
	PriceCents int             `json:"price_cents"`
	Product    *Product        `json:"product,omitempty"`
	Variant    *ProductVariant `json:"variant,omitempty"`
}

type CheckoutRequest struct {
	ShippingName      string `json:"shipping_name"`
	ShippingAddress   string `json:"shipping_address"`
	SessionID         string `json:"session_id"`
	PaymentMethod     string `json:"payment_method"`
	CouponCode        string `json:"coupon_code,omitempty"`
	CustomizationNote string `json:"customization_note,omitempty"`
}

type RazorpayVerifyRequest struct {
	OrderID   int64  `json:"order_id"`
	PaymentID string `json:"razorpay_payment_id"`
	RazorpayOrderID string `json:"razorpay_order_id"`
	Signature string `json:"razorpay_signature"`
}
