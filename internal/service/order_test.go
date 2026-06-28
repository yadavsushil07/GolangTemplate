package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
	"github.com/yadavsushil07/GolangTemplate/internal/testutil"
)

func TestOrderService_Checkout(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	ctx := context.Background()

	userRepo := repository.NewUserRepository(tdb.Pool)
	productRepo := repository.NewProductRepository(tdb.Pool)
	variantRepo := repository.NewVariantRepository(tdb.Pool)
	catRepo := repository.NewCategoryRepository(tdb.Pool)
	cartRepo := repository.NewCartRepository(tdb.Pool)
	orderRepo := repository.NewOrderRepository(tdb.Pool)
	couponRepo := repository.NewCouponRepository(tdb.Pool)
	couponSvc := service.NewCouponService(couponRepo)
	orderSvc := service.NewOrderService(tdb.Pool, orderRepo, cartRepo, productRepo, variantRepo, couponSvc)
	_ = service.NewProductService(productRepo, variantRepo, catRepo)

	t.Run("Successful COD checkout", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "cart_items", "products", "users")

		u, _ := userRepo.Create(ctx, "buyer@test.com", "customer")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "Kurti", PriceCents: 80000, Stock: 10})
		sid := "sess-order-1"
		_ = cartRepo.Upsert(ctx, sid, p.ID, nil, 2)

		order, err := orderSvc.Checkout(ctx, u.ID, model.CheckoutRequest{
			ShippingName:    "Alice",
			ShippingAddress: "12 MG Road",
			SessionID:       sid,
			PaymentMethod:   model.PaymentMethodCOD,
		})
		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, 160000, order.TotalCents)
		assert.Equal(t, model.PaymentMethodCOD, order.PaymentMethod)

		// Cart should be cleared
		items, _ := cartRepo.GetItems(ctx, sid)
		assert.Len(t, items, 0)

		// Stock should be decremented
		updatedProduct, _ := productRepo.FindByID(ctx, p.ID)
		assert.Equal(t, 8, updatedProduct.Stock)
	})

	t.Run("Checkout with coupon discount", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "cart_items", "coupons", "products", "users")

		u, _ := userRepo.Create(ctx, "buyer2@test.com", "customer")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "Suit", PriceCents: 100000, Stock: 5})
		sid := "sess-order-2"
		_ = cartRepo.Upsert(ctx, sid, p.ID, nil, 1)

		pct := 10
		couponRepo.Create(ctx, model.CreateCouponRequest{Code: "10OFF", DiscountPct: &pct})

		order, err := orderSvc.Checkout(ctx, u.ID, model.CheckoutRequest{
			ShippingName:    "Bob",
			ShippingAddress: "14 Park Ave",
			SessionID:       sid,
			PaymentMethod:   model.PaymentMethodCOD,
			CouponCode:      "10OFF",
		})
		require.NoError(t, err)
		assert.Equal(t, 90000, order.TotalCents)
		assert.Equal(t, 10000, order.DiscountCents)
		assert.Equal(t, "10OFF", order.CouponCode)
	})

	t.Run("Checkout fails on empty cart", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "cart_items", "users")
		u, _ := userRepo.Create(ctx, "empty@test.com", "customer")

		_, err := orderSvc.Checkout(ctx, u.ID, model.CheckoutRequest{
			ShippingName:    "C",
			ShippingAddress: "A",
			SessionID:       "empty-sess",
			PaymentMethod:   model.PaymentMethodCOD,
		})
		assert.Error(t, err)
	})

	t.Run("Checkout fails with insufficient stock", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "cart_items", "products", "users")
		u, _ := userRepo.Create(ctx, "stock@test.com", "customer")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "Limited", PriceCents: 10000, Stock: 1})
		sid := "sess-stock"
		_ = cartRepo.Upsert(ctx, sid, p.ID, nil, 5)

		_, err := orderSvc.Checkout(ctx, u.ID, model.CheckoutRequest{
			ShippingName:    "D",
			ShippingAddress: "E",
			SessionID:       sid,
			PaymentMethod:   model.PaymentMethodCOD,
		})
		assert.Error(t, err)
	})

	t.Run("Checkout with variant", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "cart_items", "product_variants", "products", "users")
		u, _ := userRepo.Create(ctx, "variant@test.com", "customer")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "Kurti Variant", PriceCents: 50000, Stock: 0})
		v, _ := variantRepo.Create(ctx, p.ID, model.CreateVariantRequest{Size: "M", PriceCents: 60000, Stock: 5})
		sid := "sess-variant"
		_ = cartRepo.Upsert(ctx, sid, p.ID, &v.ID, 2)

		order, err := orderSvc.Checkout(ctx, u.ID, model.CheckoutRequest{
			ShippingName:    "E",
			ShippingAddress: "F",
			SessionID:       sid,
			PaymentMethod:   model.PaymentMethodCOD,
		})
		require.NoError(t, err)
		assert.Equal(t, 120000, order.TotalCents)

		refreshed, _ := variantRepo.FindByID(ctx, v.ID)
		assert.Equal(t, 3, refreshed.Stock)
	})
}
