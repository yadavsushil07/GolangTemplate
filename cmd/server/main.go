package main

import (
	"context"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/config"
	"github.com/yadavsushil07/GolangTemplate/internal/handler"
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
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.OTPExpiryMinutes)
	productSvc := service.NewProductService(productRepo)
	cartSvc := service.NewCartService(cartRepo, productRepo)
	orderSvc := service.NewOrderService(db, orderRepo, cartRepo, productRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	productH := handler.NewProductHandler(productSvc)
	cartH := handler.NewCartHandler(cartSvc)
	orderH := handler.NewOrderHandler(orderSvc)
	vendorH := handler.NewVendorHandler(productSvc, orderSvc)

	r := router.New(authSvc, authH, productH, cartH, orderH, vendorH, cfg.RateLimitPerMinute)

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
