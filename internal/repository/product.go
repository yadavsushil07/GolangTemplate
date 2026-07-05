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
	_ = i // suppress unused warning

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
	var ids []int64
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.PriceCents, &p.ImageURL, &p.Stock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return products, nil
	}

	// Batch-load categories for all products
	catMap := make(map[int64][]model.Category)
	cRows, err := r.db.Query(ctx, `
		SELECT pc.product_id, c.id, c.name, c.slug, c.sort_order
		FROM categories c JOIN product_categories pc ON pc.category_id = c.id
		WHERE pc.product_id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer cRows.Close()
	for cRows.Next() {
		var pid int64
		var c model.Category
		if err := cRows.Scan(&pid, &c.ID, &c.Name, &c.Slug, &c.SortOrder); err != nil {
			return nil, err
		}
		catMap[pid] = append(catMap[pid], c)
	}
	if err := cRows.Err(); err != nil {
		return nil, err
	}

	// Batch-load attribute values for all products
	avMap := make(map[int64][]model.AttributeValue)
	avRows, err := r.db.Query(ctx, `
		SELECT pav.product_id, av.id, av.attribute_id, av.value, av.sort_order
		FROM attribute_values av
		JOIN product_attribute_values pav ON pav.attribute_value_id = av.id
		WHERE pav.product_id = ANY($1)
		ORDER BY av.attribute_id, av.sort_order`, ids)
	if err != nil {
		return nil, err
	}
	defer avRows.Close()
	for avRows.Next() {
		var pid int64
		var av model.AttributeValue
		if err := avRows.Scan(&pid, &av.ID, &av.AttributeID, &av.Value, &av.SortOrder); err != nil {
			return nil, err
		}
		avMap[pid] = append(avMap[pid], av)
	}
	if err := avRows.Err(); err != nil {
		return nil, err
	}

	// Batch-load variants for all products
	varMap := make(map[int64][]model.ProductVariant)
	vRows, err := r.db.Query(ctx, `
		SELECT id, product_id, size, color, price_cents, stock, sku, is_active
		FROM product_variants WHERE product_id = ANY($1) ORDER BY product_id, size`, ids)
	if err != nil {
		return nil, err
	}
	defer vRows.Close()
	for vRows.Next() {
		var v model.ProductVariant
		if err := vRows.Scan(&v.ID, &v.ProductID, &v.Size, &v.Color, &v.PriceCents, &v.Stock, &v.SKU, &v.IsActive); err != nil {
			return nil, err
		}
		varMap[v.ProductID] = append(varMap[v.ProductID], v)
	}
	if err := vRows.Err(); err != nil {
		return nil, err
	}

	// Attach relations to each product
	for i := range products {
		pid := products[i].ID
		products[i].Categories = catMap[pid]
		products[i].AttributeValues = avMap[pid]
		products[i].Variants = varMap[pid]
	}

	return products, nil
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
	if err := vRows.Err(); err != nil {
		return nil, err
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
	if err := iRows.Err(); err != nil {
		return nil, err
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
	if err := cRows.Err(); err != nil {
		return nil, err
	}

	// Load attribute values
	avRows, err := r.db.Query(ctx, `
		SELECT av.id, av.attribute_id, av.value, av.sort_order
		FROM attribute_values av
		JOIN product_attribute_values pav ON pav.attribute_value_id = av.id
		WHERE pav.product_id = $1
		ORDER BY av.attribute_id, av.sort_order`, id)
	if err != nil {
		return nil, err
	}
	defer avRows.Close()
	for avRows.Next() {
		var av model.AttributeValue
		if err := avRows.Scan(&av.ID, &av.AttributeID, &av.Value, &av.SortOrder); err != nil {
			return nil, err
		}
		p.AttributeValues = append(p.AttributeValues, av)
	}
	if err := avRows.Err(); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *ProductRepository) Create(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	base := slugify(req.Name)
	p := &model.Product{}
	err = tx.QueryRow(ctx,
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

	// Set categories
	for _, cid := range req.CategoryIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			p.ID, cid); err != nil {
			return nil, fmt.Errorf("set category %d: %w", cid, err)
		}
	}

	// Set attribute values
	for _, vid := range req.AttributeValueIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO product_attribute_values (product_id, attribute_value_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			p.ID, vid); err != nil {
			return nil, fmt.Errorf("set attribute value %d: %w", vid, err)
		}
	}

	// Create initial variants
	for _, vreq := range req.Variants {
		if _, err := tx.Exec(ctx,
			`INSERT INTO product_variants (product_id, size, color, price_cents, stock, sku)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			p.ID, vreq.Size, vreq.Color, vreq.PriceCents, vreq.Stock, vreq.SKU); err != nil {
			return nil, fmt.Errorf("create variant: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Reload fully with all relations
	return r.FindByID(ctx, p.ID)
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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Update scalar columns on the products table
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
		`UPDATE products SET %s WHERE id = $%d RETURNING id`,
		strings.Join(sets, ", "), i,
	)

	var returnedID int64
	err = tx.QueryRow(ctx, query, args...).Scan(&returnedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Replace categories if provided
	if req.CategoryIDs != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM product_categories WHERE product_id = $1`, id); err != nil {
			return nil, fmt.Errorf("clear categories: %w", err)
		}
		for _, cid := range *req.CategoryIDs {
			if _, err := tx.Exec(ctx,
				`INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				id, cid); err != nil {
				return nil, fmt.Errorf("set category %d: %w", cid, err)
			}
		}
	}

	// Replace attribute values if provided
	if req.AttributeValueIDs != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM product_attribute_values WHERE product_id = $1`, id); err != nil {
			return nil, fmt.Errorf("clear attribute values: %w", err)
		}
		for _, vid := range *req.AttributeValueIDs {
			if _, err := tx.Exec(ctx,
				`INSERT INTO product_attribute_values (product_id, attribute_value_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				id, vid); err != nil {
				return nil, fmt.Errorf("set attribute value %d: %w", vid, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Reload fully with all relations
	return r.FindByID(ctx, id)
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
