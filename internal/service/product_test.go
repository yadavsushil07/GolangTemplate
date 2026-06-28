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

func setupProductService(t *testing.T) (*service.ProductService, *testutil.TestDB) {
	tdb := testutil.NewTestDB(t)
	productRepo := repository.NewProductRepository(tdb.Pool)
	variantRepo := repository.NewVariantRepository(tdb.Pool)
	catRepo := repository.NewCategoryRepository(tdb.Pool)
	svc := service.NewProductService(productRepo, variantRepo, catRepo)
	return svc, tdb
}

func TestProductService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and GetByID", func(t *testing.T) {
		svc, tdb := setupProductService(t)
		tdb.TruncateTables(t, "products")

		p, err := svc.Create(ctx, model.CreateProductRequest{
			Name:       "Silk Saree",
			PriceCents: 100000,
			Stock:      5,
		})
		require.NoError(t, err)
		assert.Equal(t, "Silk Saree", p.Name)

		found, err := svc.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, found.ID)
	})

	t.Run("Create requires name", func(t *testing.T) {
		svc, tdb := setupProductService(t)
		tdb.TruncateTables(t, "products")

		_, err := svc.Create(ctx, model.CreateProductRequest{PriceCents: 100})
		assert.Error(t, err)
	})

	t.Run("Create requires positive price", func(t *testing.T) {
		svc, tdb := setupProductService(t)
		tdb.TruncateTables(t, "products")

		_, err := svc.Create(ctx, model.CreateProductRequest{Name: "X", PriceCents: 0})
		assert.Error(t, err)
	})

	t.Run("List active products", func(t *testing.T) {
		svc, tdb := setupProductService(t)
		tdb.TruncateTables(t, "products")

		svc.Create(ctx, model.CreateProductRequest{Name: "A", PriceCents: 100, Stock: 1})
		svc.Create(ctx, model.CreateProductRequest{Name: "B", PriceCents: 200, Stock: 0})

		products, err := svc.List(ctx, true, "")
		require.NoError(t, err)
		assert.Len(t, products, 2)
	})

	t.Run("Deactivate product", func(t *testing.T) {
		svc, tdb := setupProductService(t)
		tdb.TruncateTables(t, "products")

		p, _ := svc.Create(ctx, model.CreateProductRequest{Name: "P", PriceCents: 100, Stock: 1})
		err := svc.Deactivate(ctx, p.ID)
		require.NoError(t, err)

		active, _ := svc.List(ctx, true, "")
		assert.Len(t, active, 0)
	})

	t.Run("Add variant", func(t *testing.T) {
		svc, tdb := setupProductService(t)
		tdb.TruncateTables(t, "products")

		p, _ := svc.Create(ctx, model.CreateProductRequest{Name: "Kurti", PriceCents: 50000, Stock: 10})
		v, err := svc.AddVariant(ctx, p.ID, model.CreateVariantRequest{Size: "L", PriceCents: 55000, Stock: 3})
		require.NoError(t, err)
		assert.Equal(t, "L", v.Size)
	})

	t.Run("AddVariant requires size", func(t *testing.T) {
		svc, tdb := setupProductService(t)
		tdb.TruncateTables(t, "products")

		p, _ := svc.Create(ctx, model.CreateProductRequest{Name: "X", PriceCents: 100, Stock: 1})
		_, err := svc.AddVariant(ctx, p.ID, model.CreateVariantRequest{PriceCents: 100})
		assert.Error(t, err)
	})
}
