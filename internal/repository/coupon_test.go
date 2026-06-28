package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	"github.com/yadavsushil07/GolangTemplate/internal/testutil"
)

func TestCouponRepository(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewCouponRepository(tdb.Pool)
	ctx := context.Background()

	pct := 15

	t.Run("Create and FindByCode", func(t *testing.T) {
		tdb.TruncateTables(t, "coupons")
		c, err := repo.Create(ctx, model.CreateCouponRequest{
			Code:        "SAVE15",
			DiscountPct: &pct,
		})
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, "SAVE15", c.Code)
		assert.True(t, c.IsActive)

		found, err := repo.FindByCode(ctx, "SAVE15")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, c.ID, found.ID)
	})

	t.Run("FindByCode returns nil for unknown", func(t *testing.T) {
		found, err := repo.FindByCode(ctx, "NOPE")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Deactivate", func(t *testing.T) {
		tdb.TruncateTables(t, "coupons")
		c, _ := repo.Create(ctx, model.CreateCouponRequest{Code: "OFF10", DiscountPct: &pct})
		err := repo.Deactivate(ctx, c.ID)
		require.NoError(t, err)

		found, _ := repo.FindByCode(ctx, "OFF10")
		require.NotNil(t, found)
		assert.False(t, found.IsActive)
	})

	t.Run("List coupons", func(t *testing.T) {
		tdb.TruncateTables(t, "coupons")
		repo.Create(ctx, model.CreateCouponRequest{Code: "A", DiscountPct: &pct})
		repo.Create(ctx, model.CreateCouponRequest{Code: "B", DiscountPct: &pct})

		list, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})
}
