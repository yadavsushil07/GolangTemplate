package model

type CartItem struct {
	ID        int64           `json:"id"`
	SessionID string          `json:"session_id"`
	ProductID int64           `json:"product_id"`
	VariantID *int64          `json:"variant_id,omitempty"`
	Quantity  int             `json:"quantity"`
	Product   *Product        `json:"product,omitempty"`
	Variant   *ProductVariant `json:"variant,omitempty"`
}

type CartSummary struct {
	Items      []CartItem `json:"items"`
	TotalCents int        `json:"total_cents"`
}
