package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/yadavsushil07/GolangTemplate/internal/handler"
	"github.com/yadavsushil07/GolangTemplate/internal/middleware"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
)

func New(
	authSvc *service.AuthService,
	authH *handler.AuthHandler,
	productH *handler.ProductHandler,
	categoryH *handler.CategoryHandler,
	cartH *handler.CartHandler,
	orderH *handler.OrderHandler,
	vendorH *handler.VendorHandler,
	adminH *handler.AdminHandler,
	rateLimitPerMinute int,
	allowedOrigins []string,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.SecurityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Session-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RateLimit(rateLimitPerMinute))

	r.Get("/", handler.HomeHandler)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		// Auth — public
		r.Post("/auth/request-otp", authH.RequestOTP)
		r.Post("/auth/verify-otp", authH.VerifyOTP)

		// Products — public (supports ?category=slug)
		r.Get("/products", productH.List)
		r.Get("/products/slug/{slug}", productH.GetBySlug)
		r.Get("/products/{id}", productH.GetByID)

		// Categories — public
		r.Get("/categories", categoryH.List)

		// Coupon validation — public
		r.Post("/coupons/validate", orderH.ValidateCoupon)

		// Cart — public (session-based)
		r.Get("/cart", cartH.GetCart)
		r.Post("/cart", cartH.AddItem)
		r.Delete("/cart/{productId}", cartH.RemoveItem)

		// Customer — auth required
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc))
			r.Post("/checkout", orderH.Checkout)
			r.Get("/orders", orderH.ListMyOrders)
			r.Post("/payments/razorpay/create-order", orderH.CreateRazorpayOrder)
			r.Post("/payments/razorpay/verify", orderH.VerifyRazorpayPayment)
		})

		// Vendor — auth + vendor or admin role required
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc))
			r.Use(middleware.VendorOrAdmin)

			// Products
			r.Get("/vendor/products", vendorH.ListAllProducts)
			r.Post("/vendor/products", vendorH.CreateProduct)
			r.Put("/vendor/products/{id}", vendorH.UpdateProduct)
			r.Delete("/vendor/products/{id}", vendorH.DeactivateProduct)

			// Variants
			r.Post("/vendor/products/{id}/variants", vendorH.AddVariant)
			r.Delete("/vendor/products/{id}/variants/{variantId}", vendorH.DeleteVariant)

			// Images
			r.Post("/vendor/products/{id}/images", vendorH.AddImages)
			r.Delete("/vendor/products/{id}/images/{imageId}", vendorH.DeleteImage)

			// Product categories
			r.Put("/vendor/products/{id}/categories", vendorH.SetProductCategories)

			// Categories
			r.Get("/vendor/categories", vendorH.ListCategories)
			r.Post("/vendor/categories", vendorH.CreateCategory)
			r.Delete("/vendor/categories/{id}", vendorH.DeleteCategory)

			// Orders
			r.Get("/vendor/orders", vendorH.ListAllOrders)
			r.Put("/vendor/orders/{id}/status", vendorH.UpdateOrderStatus)

			// Coupons
			r.Get("/vendor/coupons", vendorH.ListCoupons)
			r.Post("/vendor/coupons", vendorH.CreateCoupon)
			r.Delete("/vendor/coupons/{id}", vendorH.DeactivateCoupon)
		})

		// Admin — auth + admin role only
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc))
			r.Use(middleware.AdminOnly)

			r.Get("/admin/summary", adminH.Summary)
			r.Get("/admin/users", adminH.ListUsers)
			r.Put("/admin/users/{id}/role", adminH.SetUserRole)
		})
	})

	return r
}
