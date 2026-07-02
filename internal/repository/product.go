package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, activeOnly bool, categorySlug string) ([]model.Product, error) {
	args := []any{}
	where := []string{}
	i := 1

	if activeOnly {
		where = append(where, fmt.Sprintf("p.is_active = $%d", i))
		args = append(args, true)
		i++
	}
	if categorySlug != "" {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM product_categories pc
			JOIN categories c ON c.id = pc.category_id
			WHERE pc.product_id = p.id AND c.slug = $%d)`, i))
		args = append(args, categorySlug)
		i++
	}

	query := `SELECT p.id, p.name, p.slug, p.description, p.price_cents, p.image_url, p.stock, p.is_active, p.created_at, p.updated_at FROM products p`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += ` ORDER BY p.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.PriceCents, &p.ImageURL, &p.Stock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// FindBySlug resolves a product by its SEO slug, then loads it fully.
func (r *ProductRepository) FindBySlug(ctx context.Context, slug string) (*model.Product, error) {
	var id int64
	err := r.db.QueryRow(ctx, `SELECT id FROM products WHERE slug = $1`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *ProductRepository) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	p := &model.Product{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug, description, price_cents, image_url, stock, is_active, created_at, updated_at FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.PriceCents, &p.ImageURL, &p.Stock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load variants
	vRows, err := r.db.Query(ctx,
		`SELECT id, product_id, size, color, price_cents, stock, sku, is_active FROM product_variants WHERE product_id = $1 ORDER BY size`, id)
	if err != nil {
		return nil, err
	}
	defer vRows.Close()
	for vRows.Next() {
		var v model.ProductVariant
		if err := vRows.Scan(&v.ID, &v.ProductID, &v.Size, &v.Color, &v.PriceCents, &v.Stock, &v.SKU, &v.IsActive); err != nil {
			return nil, err
		}
		p.Variants = append(p.Variants, v)
	}

	// Load images
	iRows, err := r.db.Query(ctx,
		`SELECT id, product_id, url, sort_order FROM product_images WHERE product_id = $1 ORDER BY sort_order`, id)
	if err != nil {
		return nil, err
	}
	defer iRows.Close()
	for iRows.Next() {
		var img model.ProductImage
		if err := iRows.Scan(&img.ID, &img.ProductID, &img.URL, &img.SortOrder); err != nil {
			return nil, err
		}
		p.Images = append(p.Images, img)
	}

	// Load categories
	cRows, err := r.db.Query(ctx, `
		SELECT c.id, c.name, c.slug, c.sort_order
		FROM categories c JOIN product_categories pc ON pc.category_id = c.id
		WHERE pc.product_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer cRows.Close()
	for cRows.Next() {
		var c model.Category
		if err := cRows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder); err != nil {
			return nil, err
		}
		p.Categories = append(p.Categories, c)
	}

	return p, nil
}

func (r *ProductRepository) Create(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	// Generate a URL-safe slug from the name, then guarantee uniqueness by
	// appending the row id. A provisional random slug avoids a unique-index
	// collision during the initial INSERT.
	base := slugify(req.Name)
	p := &model.Product{}
	err := r.db.QueryRow(ctx,
		`WITH ins AS (
			INSERT INTO products (name, description, price_cents, image_url, stock, slug)
			VALUES ($1, $2, $3, $4, $5, md5(random()::text))
			RETURNING id
		)
		UPDATE products p SET slug = $6 || '-' || p.id
		FROM ins WHERE p.id = ins.id
		RETURNING p.id, p.name, p.slug, p.description, p.price_cents, p.image_url, p.stock, p.is_active, p.created_at, p.updated_at`,
		req.Name, req.Description, req.PriceCents, req.ImageURL, req.Stock, base,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.PriceCents, &p.ImageURL, &p.Stock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// slugify converts an arbitrary product name into a lowercase, dash-separated
// URL slug containing only [a-z0-9] and single dashes.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		case !prevDash:
			b.WriteRune('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "product"
	}
	return out
}

func (r *ProductRepository) Update(ctx context.Context, id int64, req model.UpdateProductRequest) (*model.Product, error) {
	sets := []string{"updated_at = $1"}
	args := []any{time.Now()}
	i := 2

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i)); args = append(args, *req.Name); i++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i)); args = append(args, *req.Description); i++
	}
	if req.PriceCents != nil {
		sets = append(sets, fmt.Sprintf("price_cents = $%d", i)); args = append(args, *req.PriceCents); i++
	}
	if req.ImageURL != nil {
		sets = append(sets, fmt.Sprintf("image_url = $%d", i)); args = append(args, *req.ImageURL); i++
	}
	if req.Stock != nil {
		sets = append(sets, fmt.Sprintf("stock = $%d", i)); args = append(args, *req.Stock); i++
	}
	if req.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", i)); args = append(args, *req.IsActive); i++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE products SET %s WHERE id = $%d RETURNING id, name, slug, description, price_cents, image_url, stock, is_active, created_at, updated_at`,
		strings.Join(sets, ", "), i,
	)

	p := &model.Product{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.PriceCents, &p.ImageURL, &p.Stock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProductRepository) Deactivate(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE products SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *ProductRepository) DecrementStock(ctx context.Context, id int64, qty int) error {
	result, err := r.db.Exec(ctx,
		`UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`, qty, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for product %d", id)
	}
	return nil
}
