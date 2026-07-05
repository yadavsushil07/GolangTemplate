package model

import "time"

type Product struct {
	ID              int64            `json:"id"`
	Name            string           `json:"name"`
	Slug            string           `json:"slug"`
	Description     string           `json:"description"`
	PriceCents      int              `json:"price_cents"`
	ImageURL        string           `json:"image_url"`
	Stock           int              `json:"stock"`
	IsActive        bool             `json:"is_active"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Variants        []ProductVariant `json:"variants,omitempty"`
	Images          []ProductImage   `json:"images,omitempty"`
	Categories      []Category       `json:"categories,omitempty"`
	AttributeValues []AttributeValue `json:"attribute_values,omitempty"`
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

// Attribute defines a product attribute group (e.g. Size, Colour, Material).
type Attribute struct {
	ID        int64            `json:"id"`
	Name      string           `json:"name"`
	SortOrder int              `json:"sort_order"`
	Values    []AttributeValue `json:"values,omitempty"`
}

// AttributeValue is a concrete option within an attribute group.
type AttributeValue struct {
	ID          int64  `json:"id"`
	AttributeID int64  `json:"attribute_id"`
	Value       string `json:"value"`
	SortOrder   int    `json:"sort_order"`
}

type CreateProductRequest struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	PriceCents       int                    `json:"price_cents"`
	ImageURL         string                 `json:"image_url"`
	Stock            int                    `json:"stock"`
	CategoryIDs      []int64                `json:"category_ids"`
	AttributeValueIDs []int64               `json:"attribute_value_ids"`
	Variants         []CreateVariantRequest `json:"variants"`
}

type UpdateProductRequest struct {
	Name              *string  `json:"name"`
	Description       *string  `json:"description"`
	PriceCents        *int     `json:"price_cents"`
	ImageURL          *string  `json:"image_url"`
	Stock             *int     `json:"stock"`
	IsActive          *bool    `json:"is_active"`
	CategoryIDs       *[]int64 `json:"category_ids"`
	AttributeValueIDs *[]int64 `json:"attribute_value_ids"`
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
