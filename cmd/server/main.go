package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/config"
	"github.com/yadavsushil07/GolangTemplate/internal/handler"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	"github.com/yadavsushil07/GolangTemplate/internal/router"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	log.Println("database connected")

	runMigrations(cfg.DatabaseURL)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	variantRepo := repository.NewVariantRepository(db)
	catRepo := repository.NewCategoryRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	couponRepo := repository.NewCouponRepository(db)

	// Notification service (optional — degrades to logs if keys are missing)
	notifSvc := service.NewNotificationService(service.NotificationConfig{
		Fast2SMSKey: os.Getenv("FAST2SMS_API_KEY"),
		ResendKey:   os.Getenv("RESEND_API_KEY"),
		FromEmail:   getEnvDefault("FROM_EMAIL", "orders@aaryashop.com"),
		FromName:    getEnvDefault("FROM_NAME", "AaryaShop"),
		VendorEmail: os.Getenv("VENDOR_EMAIL"),
		VendorPhone: os.Getenv("VENDOR_PHONE"),
	})

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.OTPExpiryMinutes)
	authSvc.SetNotificationService(notifSvc)

	productSvc := service.NewProductService(productRepo, variantRepo, catRepo)
	cartSvc := service.NewCartService(cartRepo, productRepo, variantRepo)
	couponSvc := service.NewCouponService(couponRepo)
	orderSvc := service.NewOrderService(db, orderRepo, cartRepo, productRepo, variantRepo, couponSvc)
	orderSvc.SetNotificationService(notifSvc, userRepo)

	// Optional Razorpay payment service
	var paymentSvc *service.PaymentService
	rzpKeyID := os.Getenv("RAZORPAY_KEY_ID")
	rzpKeySecret := os.Getenv("RAZORPAY_KEY_SECRET")
	if rzpKeyID != "" && rzpKeySecret != "" {
		paymentSvc = service.NewPaymentService(rzpKeyID, rzpKeySecret, orderRepo)
		log.Println("Razorpay payment service enabled")
	}

	// Seed admin user(s) from env
	seedAdmin(userRepo, os.Getenv("ADMIN_PHONE"), os.Getenv("ADMIN_EMAIL"))

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	productH := handler.NewProductHandler(productSvc)
	categoryH := handler.NewCategoryHandler(catRepo)
	cartH := handler.NewCartHandler(cartSvc)
	orderH := handler.NewOrderHandler(orderSvc, couponSvc, paymentSvc)
	vendorH := handler.NewVendorHandler(productSvc, orderSvc, couponSvc, catRepo)
	adminH := handler.NewAdminHandler(userRepo, orderRepo, orderSvc)

	r := router.New(authSvc, authH, productH, categoryH, cartH, orderH, vendorH, adminH, cfg.RateLimitPerMinute, cfg.AllowedOrigins)

	addr := ":" + cfg.Port
	log.Printf("server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func runMigrations(databaseURL string) {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		log.Printf("migration setup warning: %v", err)
		return
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("migration warning: %v", err)
	} else {
		log.Println("migrations applied")
	}
}

// seedAdmin ensures admin users from ADMIN_PHONE and ADMIN_EMAIL exist on startup.
func seedAdmin(userRepo *repository.UserRepository, adminPhone, adminEmail string) {
	ctx := context.Background()

	for _, identifier := range []string{adminPhone, adminEmail} {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			continue
		}
		existing, err := userRepo.FindByIdentifier(ctx, identifier)
		if err != nil {
			log.Printf("[seedAdmin] error looking up %s: %v", identifier, err)
			continue
		}
		if existing == nil {
			isPhone := service.IsPhone(identifier)
			u, err := userRepo.UpsertWithContact(ctx, identifier, model.RoleAdmin, isPhone)
			if err != nil {
				log.Printf("[seedAdmin] failed to create admin %s: %v", identifier, err)
				continue
			}
			log.Printf("[seedAdmin] created admin user id=%d identifier=%s", u.ID, identifier)
		} else if existing.Role != model.RoleAdmin {
			if err := userRepo.SetRole(ctx, existing.ID, model.RoleAdmin); err != nil {
				log.Printf("[seedAdmin] failed to promote %s to admin: %v", identifier, err)
				continue
			}
			log.Printf("[seedAdmin] promoted existing user id=%d to admin", existing.ID)
		} else {
			log.Printf("[seedAdmin] admin already exists: %s (id=%d)", identifier, existing.ID)
		}
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
