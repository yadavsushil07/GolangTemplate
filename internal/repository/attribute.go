package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type AttributeRepository struct {
	db *pgxpool.Pool
}

func NewAttributeRepository(db *pgxpool.Pool) *AttributeRepository {
	return &AttributeRepository{db: db}
}

// ListAttributes returns all attribute groups with their values.
func (r *AttributeRepository) ListAttributes(ctx context.Context) ([]model.Attribute, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, sort_order FROM attributes ORDER BY sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attrs []model.Attribute
	for rows.Next() {
		var a model.Attribute
		if err := rows.Scan(&a.ID, &a.Name, &a.SortOrder); err != nil {
			return nil, err
		}
		attrs = append(attrs, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load values for each attribute
	for i := range attrs {
		vals, err := r.listValues(ctx, attrs[i].ID)
		if err != nil {
			return nil, err
		}
		attrs[i].Values = vals
	}
	return attrs, nil
}

func (r *AttributeRepository) listValues(ctx context.Context, attributeID int64) ([]model.AttributeValue, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, attribute_id, value, sort_order
		 FROM attribute_values WHERE attribute_id = $1
		 ORDER BY sort_order, value`, attributeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vals []model.AttributeValue
	for rows.Next() {
		var v model.AttributeValue
		if err := rows.Scan(&v.ID, &v.AttributeID, &v.Value, &v.SortOrder); err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	return vals, rows.Err()
}

// SetProductAttributeValues replaces all attribute value links for a product atomically.
func (r *AttributeRepository) SetProductAttributeValues(ctx context.Context, productID int64, valueIDs []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx,
		`DELETE FROM product_attribute_values WHERE product_id = $1`, productID); err != nil {
		return err
	}
	for _, vid := range valueIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO product_attribute_values (product_id, attribute_value_id)
			 VALUES ($1, $2) ON CONFLICT DO NOTHING`, productID, vid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// GetForProduct returns all attribute values linked to a product, grouped by attribute.
func (r *AttributeRepository) GetForProduct(ctx context.Context, productID int64) ([]model.AttributeValue, error) {
	rows, err := r.db.Query(ctx, `
		SELECT av.id, av.attribute_id, av.value, av.sort_order
		FROM attribute_values av
		JOIN product_attribute_values pav ON pav.attribute_value_id = av.id
		WHERE pav.product_id = $1
		ORDER BY av.attribute_id, av.sort_order`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vals []model.AttributeValue
	for rows.Next() {
		var v model.AttributeValue
		if err := rows.Scan(&v.ID, &v.AttributeID, &v.Value, &v.SortOrder); err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	return vals, rows.Err()
}
