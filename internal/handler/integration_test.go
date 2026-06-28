package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yadavsushil07/GolangTemplate/internal/handler"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	"github.com/yadavsushil07/GolangTemplate/internal/router"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
	"github.com/yadavsushil07/GolangTemplate/internal/testutil"
)

// ---- helpers ----

func buildRouter(t *testing.T, tdb *testutil.TestDB) http.Handler {
	t.Helper()
	userRepo := repository.NewUserRepository(tdb.Pool)
	productRepo := repository.NewProductRepository(tdb.Pool)
	variantRepo := repository.NewVariantRepository(tdb.Pool)
	catRepo := repository.NewCategoryRepository(tdb.Pool)
	cartRepo := repository.NewCartRepository(tdb.Pool)
	orderRepo := repository.NewOrderRepository(tdb.Pool)
	couponRepo := repository.NewCouponRepository(tdb.Pool)

	authSvc := service.NewAuthService(userRepo, "test-secret-key", 10)
	productSvc := service.NewProductService(productRepo, variantRepo, catRepo)
	cartSvc := service.NewCartService(cartRepo, productRepo, variantRepo)
	couponSvc := service.NewCouponService(couponRepo)
	orderSvc := service.NewOrderService(tdb.Pool, orderRepo, cartRepo, productRepo, variantRepo, couponSvc)

	authH := handler.NewAuthHandler(authSvc)
	productH := handler.NewProductHandler(productSvc)
	categoryH := handler.NewCategoryHandler(catRepo)
	cartH := handler.NewCartHandler(cartSvc)
	orderH := handler.NewOrderHandler(orderSvc, couponSvc, nil)
	vendorH := handler.NewVendorHandler(productSvc, orderSvc, couponSvc, catRepo)

	return router.New(authSvc, authH, productH, categoryH, cartH, orderH, vendorH, 100)
}

func do(r http.Handler, method, path string, body any, token, sessionID string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), v))
}

// ---- Auth ----

func TestAuthHandlers(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	r := buildRouter(t, tdb)

	t.Run("RequestOTP success", func(t *testing.T) {
		tdb.TruncateTables(t, "users")
		w := do(r, "POST", "/api/auth/request-otp", map[string]string{"identifier": "test@example.com"}, "", "")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RequestOTP missing identifier", func(t *testing.T) {
		w := do(r, "POST", "/api/auth/request-otp", map[string]string{}, "", "")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("VerifyOTP wrong code returns 401", func(t *testing.T) {
		tdb.TruncateTables(t, "users")
		do(r, "POST", "/api/auth/request-otp", map[string]string{"identifier": "a@b.com"}, "", "")
		w := do(r, "POST", "/api/auth/verify-otp", map[string]string{"identifier": "a@b.com", "otp": "000000"}, "", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- Products ----

func TestProductHandlers(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	r := buildRouter(t, tdb)

	t.Run("List empty products", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		w := do(r, "GET", "/api/products", nil, "", "")
		assert.Equal(t, http.StatusOK, w.Code)
		var products []any
		decodeBody(t, w, &products)
		assert.Len(t, products, 0)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		w := do(r, "GET", "/api/products/99999", nil, "", "")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ---- Vendor Products (need vendor JWT) ----

func TestVendorProductHandlers(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	r := buildRouter(t, tdb)

	// Create a vendor user and get token
	vendorToken := createVendorToken(t, tdb, r, "vendor@example.com")

	t.Run("Create product as vendor", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		w := do(r, "POST", "/api/vendor/products", model.CreateProductRequest{
			Name:       "Silk Kurti",
			PriceCents: 80000,
			Stock:      20,
		}, vendorToken, "")
		assert.Equal(t, http.StatusCreated, w.Code)

		var p model.Product
		decodeBody(t, w, &p)
		assert.Equal(t, "Silk Kurti", p.Name)
	})

	t.Run("Create product requires auth", func(t *testing.T) {
		w := do(r, "POST", "/api/vendor/products", model.CreateProductRequest{Name: "X", PriceCents: 100}, "", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("List vendor products", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		do(r, "POST", "/api/vendor/products", model.CreateProductRequest{Name: "P1", PriceCents: 100, Stock: 1}, vendorToken, "")
		do(r, "POST", "/api/vendor/products", model.CreateProductRequest{Name: "P2", PriceCents: 200, Stock: 1}, vendorToken, "")

		w := do(r, "GET", "/api/vendor/products", nil, vendorToken, "")
		assert.Equal(t, http.StatusOK, w.Code)
		var products []model.Product
		decodeBody(t, w, &products)
		assert.Len(t, products, 2)
	})

	t.Run("Add variant to product", func(t *testing.T) {
		tdb.TruncateTables(t, "product_variants", "products")
		create := do(r, "POST", "/api/vendor/products", model.CreateProductRequest{Name: "Kurti", PriceCents: 50000, Stock: 0}, vendorToken, "")
		var p model.Product
		decodeBody(t, create, &p)

		w := do(r, "POST", fmt.Sprintf("/api/vendor/products/%d/variants", p.ID), model.CreateVariantRequest{
			Size: "M", Color: "Red", PriceCents: 55000, Stock: 5,
		}, vendorToken, "")
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Deactivate product", func(t *testing.T) {
		tdb.TruncateTables(t, "products")
		create := do(r, "POST", "/api/vendor/products", model.CreateProductRequest{Name: "ToDelete", PriceCents: 1000, Stock: 1}, vendorToken, "")
		var p model.Product
		decodeBody(t, create, &p)

		w := do(r, "DELETE", fmt.Sprintf("/api/vendor/products/%d", p.ID), nil, vendorToken, "")
		assert.Equal(t, http.StatusOK, w.Code)

		// Not visible in public list
		list := do(r, "GET", "/api/products", nil, "", "")
		var products []model.Product
		decodeBody(t, list, &products)
		assert.Len(t, products, 0)
	})
}

// ---- Cart ----

func TestCartHandlers(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	r := buildRouter(t, tdb)

	t.Run("Get empty cart", func(t *testing.T) {
		w := do(r, "GET", "/api/cart", nil, "", "test-session-1")
		assert.Equal(t, http.StatusOK, w.Code)
		var cart model.CartSummary
		decodeBody(t, w, &cart)
		assert.Equal(t, 0, cart.TotalCents)
		assert.Len(t, cart.Items, 0)
	})

	t.Run("Add item to cart and get cart", func(t *testing.T) {
		tdb.TruncateTables(t, "cart_items", "products")
		vendorToken := createVendorToken(t, tdb, r, "v@v.com")
		create := do(r, "POST", "/api/vendor/products", model.CreateProductRequest{Name: "Shirt", PriceCents: 25000, Stock: 10}, vendorToken, "")
		var p model.Product
		decodeBody(t, create, &p)

		sid := "cart-test-session"
		w := do(r, "POST", "/api/cart", map[string]any{"product_id": p.ID, "quantity": 2}, "", sid)
		assert.Equal(t, http.StatusOK, w.Code)

		var cart model.CartSummary
		decodeBody(t, w, &cart)
		assert.Equal(t, 50000, cart.TotalCents)
		assert.Len(t, cart.Items, 1)
	})
}

// ---- Checkout ----

func TestCheckoutHandler(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	r := buildRouter(t, tdb)

	t.Run("Checkout full flow", func(t *testing.T) {
		tdb.TruncateTables(t, "order_items", "orders", "cart_items", "products", "users")

		vendorToken := createVendorToken(t, tdb, r, "vendor2@v.com")
		create := do(r, "POST", "/api/vendor/products", model.CreateProductRequest{Name: "Saree", PriceCents: 150000, Stock: 5}, vendorToken, "")
		var p model.Product
		decodeBody(t, create, &p)

		customerToken := createCustomerToken(t, tdb, r, "customer@c.com")
		sid := "checkout-session"

		// Add to cart
		addW := do(r, "POST", "/api/cart", map[string]any{"product_id": p.ID, "quantity": 1}, "", sid)
		assert.Equal(t, http.StatusOK, addW.Code)

		// Checkout
		w := do(r, "POST", "/api/checkout", model.CheckoutRequest{
			ShippingName:    "Alice",
			ShippingAddress: "12 MG Rd, Mumbai",
			SessionID:       sid,
			PaymentMethod:   "cod",
		}, customerToken, sid)
		assert.Equal(t, http.StatusCreated, w.Code)

		var order model.Order
		decodeBody(t, w, &order)
		assert.Equal(t, 150000, order.TotalCents)
		assert.Equal(t, "placed", order.Status)
	})

	t.Run("Checkout requires auth", func(t *testing.T) {
		w := do(r, "POST", "/api/checkout", model.CheckoutRequest{
			ShippingName: "X", ShippingAddress: "Y", SessionID: "s",
		}, "", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- Coupons ----

func TestCouponHandlers(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	r := buildRouter(t, tdb)

	vendorToken := createVendorToken(t, tdb, r, "vndcoup@v.com")

	t.Run("Vendor creates coupon", func(t *testing.T) {
		tdb.TruncateTables(t, "coupons")
		pct := 20
		w := do(r, "POST", "/api/vendor/coupons", model.CreateCouponRequest{
			Code:        "TWENTY",
			DiscountPct: &pct,
		}, vendorToken, "")
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Public coupon validation success", func(t *testing.T) {
		tdb.TruncateTables(t, "coupons")
		pct := 10
		do(r, "POST", "/api/vendor/coupons", model.CreateCouponRequest{Code: "VALIDATE10", DiscountPct: &pct}, vendorToken, "")

		w := do(r, "POST", "/api/coupons/validate", model.ValidateCouponRequest{Code: "VALIDATE10", OrderTotalCents: 50000}, "", "")
		assert.Equal(t, http.StatusOK, w.Code)

		var resp model.ValidateCouponResponse
		decodeBody(t, w, &resp)
		assert.True(t, resp.Valid)
		assert.Equal(t, 5000, resp.DiscountCents)
	})

	t.Run("Coupon validation for unknown code", func(t *testing.T) {
		tdb.TruncateTables(t, "coupons")
		w := do(r, "POST", "/api/coupons/validate", model.ValidateCouponRequest{Code: "NOGOOD", OrderTotalCents: 10000}, "", "")
		assert.Equal(t, http.StatusOK, w.Code)
		var resp model.ValidateCouponResponse
		decodeBody(t, w, &resp)
		assert.False(t, resp.Valid)
	})
}

// ---- Categories ----

func TestCategoryHandlers(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	r := buildRouter(t, tdb)
	vendorToken := createVendorToken(t, tdb, r, "vndcat@v.com")

	t.Run("Vendor creates category and public lists it", func(t *testing.T) {
		tdb.TruncateTables(t, "categories")
		w := do(r, "POST", "/api/vendor/categories", model.CreateCategoryRequest{Name: "Sarees", Slug: "sarees"}, vendorToken, "")
		assert.Equal(t, http.StatusCreated, w.Code)

		list := do(r, "GET", "/api/categories", nil, "", "")
		assert.Equal(t, http.StatusOK, list.Code)
		var cats []model.Category
		decodeBody(t, list, &cats)
		assert.Len(t, cats, 1)
		assert.Equal(t, "Sarees", cats[0].Name)
	})
}

// ---- Helpers to obtain tokens ----

func createVendorToken(t *testing.T, tdb *testutil.TestDB, _ http.Handler, email string) string {
	t.Helper()
	ctx := context.Background()
	userRepo := repository.NewUserRepository(tdb.Pool)
	u, err := userRepo.Create(ctx, email, "vendor")
	require.NoError(t, err)

	authSvc := service.NewAuthService(userRepo, "test-secret-key", 10)
	token, err := authSvc.IssueTokenForUser(ctx, u.ID)
	require.NoError(t, err)
	return token
}

func createCustomerToken(t *testing.T, tdb *testutil.TestDB, _ http.Handler, email string) string {
	t.Helper()
	ctx := context.Background()
	userRepo := repository.NewUserRepository(tdb.Pool)
	u, err := userRepo.Create(ctx, email, "customer")
	require.NoError(t, err)

	authSvc := service.NewAuthService(userRepo, "test-secret-key", 10)
	token, err := authSvc.IssueTokenForUser(ctx, u.ID)
	require.NoError(t, err)
	return token
}
