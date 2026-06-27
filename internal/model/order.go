package model

import "time"

const (
	OrderStatusPlaced    = "placed"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

type Order struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	TotalCents      int         `json:"total_cents"`
	Status          string      `json:"status"`
	ShippingName    string      `json:"shipping_name"`
	ShippingAddress string      `json:"shipping_address"`
	CreatedAt       time.Time   `json:"created_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID         int64    `json:"id"`
	OrderID    int64    `json:"order_id"`
	ProductID  int64    `json:"product_id"`
	Quantity   int      `json:"quantity"`
	PriceCents int      `json:"price_cents"`
	Product    *Product `json:"product,omitempty"`
}

type CheckoutRequest struct {
	ShippingName    string `json:"shipping_name"`
	ShippingAddress string `json:"shipping_address"`
	SessionID       string `json:"session_id"`
}
