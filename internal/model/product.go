package model

import "time"

type Product struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	PriceCents  int              `json:"price_cents"`
	ImageURL    string           `json:"image_url"`
	Stock       int              `json:"stock"`
	IsActive    bool             `json:"is_active"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Variants    []ProductVariant `json:"variants,omitempty"`
	Images      []ProductImage   `json:"images,omitempty"`
	Categories  []Category       `json:"categories,omitempty"`
}

type ProductVariant struct {
	ID         int64  `json:"id"`
	ProductID  int64  `json:"product_id"`
	Size       string `json:"size"`
	Color      string `json:"color"`
	PriceCents int    `json:"price_cents"`
	Stock      int    `json:"stock"`
	SKU        string `json:"sku"`
	IsActive   bool   `json:"is_active"`
}

type ProductImage struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
}

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int    `json:"price_cents"`
	ImageURL    string `json:"image_url"`
	Stock       int    `json:"stock"`
}

type UpdateProductRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	PriceCents  *int    `json:"price_cents"`
	ImageURL    *string `json:"image_url"`
	Stock       *int    `json:"stock"`
	IsActive    *bool   `json:"is_active"`
}

type CreateVariantRequest struct {
	Size       string `json:"size"`
	Color      string `json:"color"`
	PriceCents int    `json:"price_cents"`
	Stock      int    `json:"stock"`
	SKU        string `json:"sku"`
}

type UpdateVariantRequest struct {
	Size       *string `json:"size"`
	Color      *string `json:"color"`
	PriceCents *int    `json:"price_cents"`
	Stock      *int    `json:"stock"`
	IsActive   *bool   `json:"is_active"`
}

type CreateCategoryRequest struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
}
