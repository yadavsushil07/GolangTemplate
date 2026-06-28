package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
	"github.com/yadavsushil07/GolangTemplate/internal/testutil"
)

func setupCouponService(t *testing.T) (*service.CouponService, *testutil.TestDB) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewCouponRepository(tdb.Pool)
	return service.NewCouponService(repo), tdb
}

func TestCouponService(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid percentage coupon", func(t *testing.T) {
		svc, tdb := setupCouponService(t)
		tdb.TruncateTables(t, "coupons")
		pct := 20
		svc.Create(ctx, model.CreateCouponRequest{Code: "SAVE20", DiscountPct: &pct})

		resp, err := svc.Validate(ctx, model.ValidateCouponRequest{Code: "SAVE20", OrderTotalCents: 100000})
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 20000, resp.DiscountCents)
	})

	t.Run("Valid flat coupon", func(t *testing.T) {
		svc, tdb := setupCouponService(t)
		tdb.TruncateTables(t, "coupons")
		flat := 5000
		svc.Create(ctx, model.CreateCouponRequest{Code: "FLAT50", DiscountCents: &flat})

		resp, err := svc.Validate(ctx, model.ValidateCouponRequest{Code: "FLAT50", OrderTotalCents: 30000})
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 5000, resp.DiscountCents)
	})

	t.Run("Flat coupon capped at order total", func(t *testing.T) {
		svc, tdb := setupCouponService(t)
		tdb.TruncateTables(t, "coupons")
		flat := 50000
		svc.Create(ctx, model.CreateCouponRequest{Code: "BIG", DiscountCents: &flat})

		resp, err := svc.Validate(ctx, model.ValidateCouponRequest{Code: "BIG", OrderTotalCents: 10000})
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 10000, resp.DiscountCents)
	})

	t.Run("Minimum order not met", func(t *testing.T) {
		svc, tdb := setupCouponService(t)
		tdb.TruncateTables(t, "coupons")
		pct := 10
		svc.Create(ctx, model.CreateCouponRequest{Code: "MIN", DiscountPct: &pct, MinOrderCents: 50000})

		resp, err := svc.Validate(ctx, model.ValidateCouponRequest{Code: "MIN", OrderTotalCents: 10000})
		require.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("Expired coupon", func(t *testing.T) {
		svc, tdb := setupCouponService(t)
		tdb.TruncateTables(t, "coupons")
		pct := 5
		past := time.Now().Add(-24 * time.Hour)
		svc.Create(ctx, model.CreateCouponRequest{Code: "EXP", DiscountPct: &pct, ExpiresAt: &past})

		resp, err := svc.Validate(ctx, model.ValidateCouponRequest{Code: "EXP", OrderTotalCents: 10000})
		require.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("Invalid code", func(t *testing.T) {
		svc, tdb := setupCouponService(t)
		tdb.TruncateTables(t, "coupons")

		resp, err := svc.Validate(ctx, model.ValidateCouponRequest{Code: "GHOST", OrderTotalCents: 10000})
		require.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("Create requires discount_pct or discount_cents", func(t *testing.T) {
		svc, tdb := setupCouponService(t)
		tdb.TruncateTables(t, "coupons")
		_, err := svc.Create(ctx, model.CreateCouponRequest{Code: "BAD"})
		assert.Error(t, err)
	})
}
