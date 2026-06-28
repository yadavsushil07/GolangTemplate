package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type CartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetItems(ctx context.Context, sessionID string) ([]model.CartItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ci.id, ci.session_id, ci.product_id, ci.variant_id, ci.quantity,
		       p.id, p.name, p.description, p.price_cents, p.image_url, p.stock, p.is_active, p.created_at, p.updated_at,
		       v.id, v.product_id, v.size, v.color, v.price_cents, v.stock, v.sku, v.is_active
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		LEFT JOIN product_variants v ON v.id = ci.variant_id
		WHERE ci.session_id = $1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CartItem
	for rows.Next() {
		var ci model.CartItem
		var p model.Product
		var v model.ProductVariant
		var vID *int64
		var vProductID *int64
		var vSize, vColor, vSKU *string
		var vPriceCents, vStock *int
		var vIsActive *bool

		if err := rows.Scan(
			&ci.ID, &ci.SessionID, &ci.ProductID, &ci.VariantID, &ci.Quantity,
			&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.ImageURL, &p.Stock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&vID, &vProductID, &vSize, &vColor, &vPriceCents, &vStock, &vSKU, &vIsActive,
		); err != nil {
			return nil, err
		}
		ci.Product = &p
		if vID != nil {
			v.ID = *vID
			v.ProductID = *vProductID
			v.Size = *vSize
			v.PriceCents = *vPriceCents
			v.Stock = *vStock
			v.IsActive = *vIsActive
			if vColor != nil {
				v.Color = *vColor
			}
			if vSKU != nil {
				v.SKU = *vSKU
			}
			ci.Variant = &v
		}
		items = append(items, ci)
	}
	return items, rows.Err()
}

func (r *CartRepository) Upsert(ctx context.Context, sessionID string, productID int64, variantID *int64, qty int) error {
	if variantID != nil {
		_, err := r.db.Exec(ctx, `
			INSERT INTO cart_items (session_id, product_id, variant_id, quantity)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (session_id, variant_id)
			DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity`,
			sessionID, productID, *variantID, qty)
		return err
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO cart_items (session_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (session_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity`,
		sessionID, productID, qty)
	return err
}

func (r *CartRepository) Remove(ctx context.Context, sessionID string, productID int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM cart_items WHERE session_id = $1 AND product_id = $2`,
		sessionID, productID)
	return err
}

func (r *CartRepository) RemoveVariant(ctx context.Context, sessionID string, variantID int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM cart_items WHERE session_id = $1 AND variant_id = $2`,
		sessionID, variantID)
	return err
}

func (r *CartRepository) Clear(ctx context.Context, sessionID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM cart_items WHERE session_id = $1`, sessionID)
	return err
}
