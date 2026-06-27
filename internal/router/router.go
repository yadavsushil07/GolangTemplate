package router

import (
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
	cartH *handler.CartHandler,
	orderH *handler.OrderHandler,
	vendorH *handler.VendorHandler,
	rateLimitPerMinute int,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Session-ID"},
		AllowCredentials: true,
	}))
	r.Use(middleware.RateLimit(rateLimitPerMinute))

	r.Route("/api", func(r chi.Router) {
		// Auth - public
		r.Post("/auth/request-otp", authH.RequestOTP)
		r.Post("/auth/verify-otp", authH.VerifyOTP)

		// Products - public
		r.Get("/products", productH.List)
		r.Get("/products/{id}", productH.GetByID)

		// Cart - public (session-based)
		r.Get("/cart", cartH.GetCart)
		r.Post("/cart", cartH.AddItem)
		r.Delete("/cart/{productId}", cartH.RemoveItem)

		// Customer - auth required
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc))
			r.Post("/checkout", orderH.Checkout)
			r.Get("/orders", orderH.ListMyOrders)
		})

		// Vendor - auth + vendor role required
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc))
			r.Use(middleware.VendorOnly)
			r.Get("/vendor/products", vendorH.ListAllProducts)
			r.Post("/vendor/products", vendorH.CreateProduct)
			r.Put("/vendor/products/{id}", vendorH.UpdateProduct)
			r.Delete("/vendor/products/{id}", vendorH.DeactivateProduct)
			r.Get("/vendor/orders", vendorH.ListAllOrders)
			r.Put("/vendor/orders/{id}/status", vendorH.UpdateOrderStatus)
		})
	})

	return r
}
