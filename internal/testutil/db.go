// Package testutil provides helpers for integration tests using real Postgres via testcontainers.
package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	pgmodule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB holds a pool and the container for teardown.
type TestDB struct {
	Pool      *pgxpool.Pool
	container testcontainers.Container
}

// NewTestDB starts a Postgres container, applies all migrations, and returns a TestDB.
// The pool is closed and the container stopped when t.Cleanup runs.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := pgmodule.Run(ctx,
		"postgres:16-alpine",
		pgmodule.WithDatabase("testdb"),
		pgmodule.WithUsername("testuser"),
		pgmodule.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	// Wait for ready
	for i := 0; i < 10; i++ {
		if err := pool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Run migrations from the repo root migrations/ folder
	m, err := migrate.New("file://../../migrations", connStr)
	if err != nil {
		t.Fatalf("migrate new: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}
	fmt.Println("migrations applied to test DB")

	tdb := &TestDB{Pool: pool, container: pgContainer}
	t.Cleanup(func() {
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
	})
	return tdb
}

// TruncateTables resets tables between tests for isolation.
func (tdb *TestDB) TruncateTables(t *testing.T, tables ...string) {
	t.Helper()
	ctx := context.Background()
	for _, tbl := range tables {
		if _, err := tdb.Pool.Exec(ctx, "TRUNCATE TABLE "+tbl+" RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
}
