package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type VariantRepository struct {
	db *pgxpool.Pool
}

func NewVariantRepository(db *pgxpool.Pool) *VariantRepository {
	return &VariantRepository{db: db}
}

func (r *VariantRepository) ListForProduct(ctx context.Context, productID int64) ([]model.ProductVariant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, size, color, price_cents, stock, sku, is_active
		FROM product_variants WHERE product_id = $1 ORDER BY size, color`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []model.ProductVariant
	for rows.Next() {
		var v model.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.Size, &v.Color, &v.PriceCents, &v.Stock, &v.SKU, &v.IsActive); err != nil {
			return nil, err
		}
		variants = append(variants, v)
	}
	return variants, rows.Err()
}

func (r *VariantRepository) FindByID(ctx context.Context, id int64) (*model.ProductVariant, error) {
	v := &model.ProductVariant{}
	err := r.db.QueryRow(ctx, `
		SELECT id, product_id, size, color, price_cents, stock, sku, is_active
		FROM product_variants WHERE id = $1`, id,
	).Scan(&v.ID, &v.ProductID, &v.Size, &v.Color, &v.PriceCents, &v.Stock, &v.SKU, &v.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return v, err
}

func (r *VariantRepository) Create(ctx context.Context, productID int64, req model.CreateVariantRequest) (*model.ProductVariant, error) {
	v := &model.ProductVariant{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO product_variants (product_id, size, color, price_cents, stock, sku)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, product_id, size, color, price_cents, stock, sku, is_active`,
		productID, req.Size, req.Color, req.PriceCents, req.Stock, req.SKU,
	).Scan(&v.ID, &v.ProductID, &v.Size, &v.Color, &v.PriceCents, &v.Stock, &v.SKU, &v.IsActive)
	return v, err
}

func (r *VariantRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM product_variants WHERE id = $1`, id)
	return err
}

func (r *VariantRepository) DecrementStock(ctx context.Context, id int64, qty int) error {
	result, err := r.db.Exec(ctx,
		`UPDATE product_variants SET stock = stock - $1 WHERE id = $2 AND stock >= $1`, qty, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for variant %d", id)
	}
	return nil
}

func (r *VariantRepository) AddImages(ctx context.Context, productID int64, urls []string) error {
	for i, url := range urls {
		_, err := r.db.Exec(ctx,
			`INSERT INTO product_images (product_id, url, sort_order) VALUES ($1, $2, $3)`,
			productID, url, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *VariantRepository) GetImages(ctx context.Context, productID int64) ([]model.ProductImage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, url, sort_order FROM product_images
		WHERE product_id = $1 ORDER BY sort_order`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imgs []model.ProductImage
	for rows.Next() {
		var img model.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.URL, &img.SortOrder); err != nil {
			return nil, err
		}
		imgs = append(imgs, img)
	}
	return imgs, rows.Err()
}

func (r *VariantRepository) DeleteImage(ctx context.Context, imageID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM product_images WHERE id = $1`, imageID)
	return err
}
