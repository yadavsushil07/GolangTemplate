package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	u := &model.User{}
	var phone, email *string
	err := row.Scan(&u.ID, &u.Identifier, &u.Role, &phone, &email, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	if phone != nil {
		u.Phone = *phone
	}
	if email != nil {
		u.Email = *email
	}
	return u, nil
}

const userCols = `id, identifier, role, phone, email, created_at`

func (r *UserRepository) FindByIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE identifier = $1`, identifier))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) Create(ctx context.Context, identifier, role string) (*model.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`INSERT INTO users (identifier, role) VALUES ($1, $2)
		 RETURNING `+userCols,
		identifier, role))
	return u, err
}

// UpsertWithContact creates or updates a user, setting phone/email based on identifier type.
func (r *UserRepository) UpsertWithContact(ctx context.Context, identifier, role string, isPhone bool) (*model.User, error) {
	var q string
	if isPhone {
		q = `INSERT INTO users (identifier, role, phone)
			 VALUES ($1, $2, $1)
			 ON CONFLICT (identifier) DO UPDATE SET phone = EXCLUDED.phone
			 RETURNING ` + userCols
	} else {
		q = `INSERT INTO users (identifier, role, email)
			 VALUES ($1, $2, $1)
			 ON CONFLICT (identifier) DO UPDATE SET email = EXCLUDED.email
			 RETURNING ` + userCols
	}
	u, err := scanUser(r.db.QueryRow(ctx, q, identifier, role))
	return u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) SetRole(ctx context.Context, id int64, role string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, roleFilter string, limit, offset int) ([]model.User, error) {
	var rows pgx.Rows
	var err error
	if roleFilter != "" {
		rows, err = r.db.Query(ctx,
			`SELECT `+userCols+` FROM users WHERE role = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			roleFilter, limit, offset)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT `+userCols+` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}
