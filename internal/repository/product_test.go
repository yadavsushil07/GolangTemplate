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

func TestProductRepository(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	productRepo := repository.NewProductRepository(tdb.Pool)
	variantRepo := repository.NewVariantRepository(tdb.Pool)
	catRepo := repository.NewCategoryRepository(tdb.Pool)
	ctx := context.Background()

	t.Run("Create and FindByID", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		p, err := productRepo.Create(ctx, model.CreateProductRequest{
			Name:        "Silk Saree",
			Description: "Handwoven silk",
			PriceCents:  250000,
			ImageURL:    "https://example.com/saree.jpg",
			Stock:       10,
		})
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, "Silk Saree", p.Name)

		found, err := productRepo.FindByID(ctx, p.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, p.ID, found.ID)
	})

	t.Run("List active only", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		p1, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "P1", PriceCents: 100, Stock: 1})
		p2, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "P2", PriceCents: 200, Stock: 0})
		require.NotNil(t, p1)
		require.NotNil(t, p2)

		// deactivate p2
		_ = productRepo.Deactivate(ctx, p2.ID)

		active, err := productRepo.List(ctx, true, "")
		require.NoError(t, err)
		assert.Len(t, active, 1)
		assert.Equal(t, "P1", active[0].Name)
	})

	t.Run("List by category slug", func(t *testing.T) {
		tdb.TruncateTables(t, "categories", "product_categories", "products")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "P Cat", PriceCents: 100, Stock: 5})
		cat, err := catRepo.Create(ctx, model.CreateCategoryRequest{Name: "Sarees", Slug: "sarees"})
		require.NoError(t, err)
		_ = catRepo.SetProductCategories(ctx, p.ID, []int64{cat.ID})

		filtered, err := productRepo.List(ctx, true, "sarees")
		require.NoError(t, err)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "P Cat", filtered[0].Name)

		noMatch, err := productRepo.List(ctx, true, "lehengas")
		require.NoError(t, err)
		assert.Len(t, noMatch, 0)
	})

	t.Run("Update product", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "Old", PriceCents: 100, Stock: 1})
		newName := "New Name"
		updated, err := productRepo.Update(ctx, p.ID, model.UpdateProductRequest{Name: &newName})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "New Name", updated.Name)
	})

	t.Run("DecrementStock", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "S", PriceCents: 100, Stock: 5})
		err := productRepo.DecrementStock(ctx, p.ID, 3)
		require.NoError(t, err)

		found, _ := productRepo.FindByID(ctx, p.ID)
		assert.Equal(t, 2, found.Stock)
	})

	t.Run("Variant CRUD", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "Kurti", PriceCents: 50000, Stock: 10})

		v, err := variantRepo.Create(ctx, p.ID, model.CreateVariantRequest{
			Size: "M", Color: "Red", PriceCents: 55000, Stock: 5,
		})
		require.NoError(t, err)
		assert.Equal(t, "M", v.Size)
		assert.Equal(t, 55000, v.PriceCents)

		variants, err := variantRepo.ListForProduct(ctx, p.ID)
		require.NoError(t, err)
		assert.Len(t, variants, 1)

		// FindByID
		found, err := variantRepo.FindByID(ctx, v.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, v.ID, found.ID)

		// Decrement variant stock
		err = variantRepo.DecrementStock(ctx, v.ID, 3)
		require.NoError(t, err)
		refreshed, _ := variantRepo.FindByID(ctx, v.ID)
		assert.Equal(t, 2, refreshed.Stock)

		// Delete variant
		err = variantRepo.Delete(ctx, v.ID)
		require.NoError(t, err)
		afterDelete, _ := variantRepo.ListForProduct(ctx, p.ID)
		assert.Len(t, afterDelete, 0)
	})

	t.Run("Product images", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		p, _ := productRepo.Create(ctx, model.CreateProductRequest{Name: "Dress", PriceCents: 10000, Stock: 3})
		err := variantRepo.AddImages(ctx, p.ID, []string{"https://a.com/1.jpg", "https://a.com/2.jpg"})
		require.NoError(t, err)

		imgs, err := variantRepo.GetImages(ctx, p.ID)
		require.NoError(t, err)
		assert.Len(t, imgs, 2)
		assert.Equal(t, "https://a.com/1.jpg", imgs[0].URL)

		err = variantRepo.DeleteImage(ctx, p.ID, imgs[0].ID)
		require.NoError(t, err)
		imgs2, _ := variantRepo.GetImages(ctx, p.ID)
		assert.Len(t, imgs2, 1)
	})
}
