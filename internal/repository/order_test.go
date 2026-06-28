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

func TestOrderRepository(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(tdb.Pool)
	productRepo := repository.NewProductRepository(tdb.Pool)
	orderRepo := repository.NewOrderRepository(tdb.Pool)
	ctx := context.Background()

	t.Run("Create order and list by user", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "products", "users")

		u, err := userRepo.Create(ctx, "buyer@example.com", "customer")
		require.NoError(t, err)

		p, err := productRepo.Create(ctx, model.CreateProductRequest{Name: "Book", PriceCents: 50000, Stock: 5})
		require.NoError(t, err)

		order, err := orderRepo.Create(ctx, &model.Order{
			UserID:          u.ID,
			TotalCents:      50000,
			DiscountCents:   0,
			PaymentMethod:   model.PaymentMethodCOD,
			PaymentStatus:   model.PaymentStatusPending,
			ShippingName:    "Test User",
			ShippingAddress: "123 Main St",
		})
		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, u.ID, order.UserID)
		assert.Equal(t, "placed", order.Status)

		err = orderRepo.AddItem(ctx, order.ID, p.ID, nil, 1, 50000)
		require.NoError(t, err)

		orders, err := orderRepo.ListByUser(ctx, u.ID)
		require.NoError(t, err)
		assert.Len(t, orders, 1)
		assert.Equal(t, order.ID, orders[0].ID)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "users")
		u, _ := userRepo.Create(ctx, "buyer2@example.com", "customer")
		order, err := orderRepo.Create(ctx, &model.Order{
			UserID:          u.ID,
			TotalCents:      10000,
			PaymentMethod:   model.PaymentMethodCOD,
			PaymentStatus:   model.PaymentStatusPending,
			ShippingName:    "U",
			ShippingAddress: "A",
		})
		require.NoError(t, err)

		err = orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusShipped)
		require.NoError(t, err)

		found, err := orderRepo.FindByID(ctx, order.ID)
		require.NoError(t, err)
		assert.Equal(t, model.OrderStatusShipped, found.Status)
	})

	t.Run("UpdatePaymentStatus", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "users")
		u, _ := userRepo.Create(ctx, "buyer3@example.com", "customer")
		order, err := orderRepo.Create(ctx, &model.Order{
			UserID:          u.ID,
			TotalCents:      20000,
			PaymentMethod:   model.PaymentMethodRazorpay,
			PaymentStatus:   model.PaymentStatusPending,
			ShippingName:    "U",
			ShippingAddress: "A",
		})
		require.NoError(t, err)

		_ = orderRepo.UpdateRazorpayOrderID(ctx, order.ID, "rzp_order_abc123")
		_ = orderRepo.UpdatePaymentStatus(ctx, order.ID, model.PaymentStatusPaid)

		found, err := orderRepo.FindByID(ctx, order.ID)
		require.NoError(t, err)
		assert.Equal(t, model.PaymentStatusPaid, found.PaymentStatus)
		assert.Equal(t, "rzp_order_abc123", found.RazorpayOrderID)
	})

	t.Run("ListAll returns all orders", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "users")
		u, _ := userRepo.Create(ctx, "a@a.com", "customer")
		for i := 0; i < 3; i++ {
			_, err := orderRepo.Create(ctx, &model.Order{
				UserID: u.ID, TotalCents: 1000,
				PaymentMethod: model.PaymentMethodCOD, PaymentStatus: model.PaymentStatusPending,
				ShippingName: "N", ShippingAddress: "A",
			})
			require.NoError(t, err)
		}
		all, err := orderRepo.ListAll(ctx)
		require.NoError(t, err)
		assert.Len(t, all, 3)
	})
}
