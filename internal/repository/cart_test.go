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

func TestCartRepository(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	productRepo := repository.NewProductRepository(tdb.Pool)
	cartRepo := repository.NewCartRepository(tdb.Pool)
	ctx := context.Background()

	setup := func(t *testing.T) int64 {
		t.Helper()
		tdb.TruncateTables(t, "cart_items", "products")
		p, err := productRepo.Create(ctx, model.CreateProductRequest{
			Name: "Test Product", PriceCents: 10000, Stock: 20,
		})
		require.NoError(t, err)
		return p.ID
	}

	t.Run("Upsert and GetItems without variant", func(t *testing.T) {
		productID := setup(t)
		sid := "sess-1"
		err := cartRepo.Upsert(ctx, sid, productID, nil, 2)
		require.NoError(t, err)

		items, err := cartRepo.GetItems(ctx, sid)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, 2, items[0].Quantity)
		assert.NotNil(t, items[0].Product)
		assert.Nil(t, items[0].Variant)
	})

	t.Run("Upsert increases quantity on re-add", func(t *testing.T) {
		productID := setup(t)
		sid := "sess-2"
		_ = cartRepo.Upsert(ctx, sid, productID, nil, 1)
		_ = cartRepo.Upsert(ctx, sid, productID, nil, 2)

		items, _ := cartRepo.GetItems(ctx, sid)
		assert.Equal(t, 3, items[0].Quantity)
	})

	t.Run("Remove item", func(t *testing.T) {
		productID := setup(t)
		sid := "sess-3"
		_ = cartRepo.Upsert(ctx, sid, productID, nil, 1)
		err := cartRepo.Remove(ctx, sid, productID)
		require.NoError(t, err)

		items, _ := cartRepo.GetItems(ctx, sid)
		assert.Len(t, items, 0)
	})

	t.Run("Clear cart", func(t *testing.T) {
		productID := setup(t)
		sid := "sess-4"
		_ = cartRepo.Upsert(ctx, sid, productID, nil, 5)
		err := cartRepo.Clear(ctx, sid)
		require.NoError(t, err)

		items, _ := cartRepo.GetItems(ctx, sid)
		assert.Len(t, items, 0)
	})

	t.Run("GetItems for unknown session is empty", func(t *testing.T) {
		items, err := cartRepo.GetItems(ctx, "unknown-session")
		require.NoError(t, err)
		assert.Len(t, items, 0)
	})
}
