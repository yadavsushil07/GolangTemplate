package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, slug, sort_order FROM categories ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *CategoryRepository) Create(ctx context.Context, req model.CreateCategoryRequest) (*model.Category, error) {
	c := &model.Category{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO categories (name, slug, sort_order) VALUES ($1, $2, $3)
		 RETURNING id, name, slug, sort_order`,
		req.Name, req.Slug, req.SortOrder,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder)
	return c, err
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	c := &model.Category{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug, sort_order FROM categories WHERE slug = $1`, slug,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *CategoryRepository) SetProductCategories(ctx context.Context, productID int64, categoryIDs []int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM product_categories WHERE product_id = $1`, productID)
	if err != nil {
		return err
	}
	for _, cid := range categoryIDs {
		_, err = r.db.Exec(ctx,
			`INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`, productID, cid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CategoryRepository) GetForProduct(ctx context.Context, productID int64) ([]model.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.name, c.slug, c.sort_order
		FROM categories c
		JOIN product_categories pc ON pc.category_id = c.id
		WHERE pc.product_id = $1
		ORDER BY c.sort_order`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}
