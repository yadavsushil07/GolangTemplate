package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type OrderService struct {
	db          *pgxpool.Pool
	orderRepo   *repository.OrderRepository
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
	variantRepo *repository.VariantRepository
	couponSvc   *CouponService
	notif       *NotificationService
	userRepo    *repository.UserRepository
}

func NewOrderService(
	db *pgxpool.Pool,
	orderRepo *repository.OrderRepository,
	cartRepo *repository.CartRepository,
	productRepo *repository.ProductRepository,
	variantRepo *repository.VariantRepository,
	couponSvc *CouponService,
) *OrderService {
	return &OrderService{
		db: db, orderRepo: orderRepo, cartRepo: cartRepo,
		productRepo: productRepo, variantRepo: variantRepo, couponSvc: couponSvc,
	}
}

// SetNotificationService wires in notifications after construction.
func (s *OrderService) SetNotificationService(n *NotificationService, userRepo *repository.UserRepository) {
	s.notif = n
	s.userRepo = userRepo
}

func (s *OrderService) Checkout(ctx context.Context, userID int64, req model.CheckoutRequest) (*model.Order, error) {
	if req.ShippingName == "" || req.ShippingAddress == "" {
		return nil, fmt.Errorf("shipping name and address are required")
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = model.PaymentMethodCOD
	}
	if paymentMethod != model.PaymentMethodCOD && paymentMethod != model.PaymentMethodRazorpay {
		return nil, fmt.Errorf("invalid payment method: %s", paymentMethod)
	}

	items, err := s.cartRepo.GetItems(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	subtotalCents := 0
	for _, item := range items {
		if item.Variant != nil {
			result, err := tx.Exec(ctx,
				`UPDATE product_variants SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
				item.Quantity, item.Variant.ID)
			if err != nil {
				return nil, err
			}
			if result.RowsAffected() == 0 {
				return nil, fmt.Errorf("insufficient stock for variant: %s %s", item.Product.Name, item.Variant.Size)
			}
			subtotalCents += item.Variant.PriceCents * item.Quantity
		} else {
			if item.Product == nil {
				return nil, fmt.Errorf("product data missing in cart")
			}
			result, err := tx.Exec(ctx,
				`UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
				item.Quantity, item.ProductID)
			if err != nil {
				return nil, err
			}
			if result.RowsAffected() == 0 {
				return nil, fmt.Errorf("insufficient stock for product: %s", item.Product.Name)
			}
			subtotalCents += item.Product.PriceCents * item.Quantity
		}
	}

	// Apply coupon
	discountCents := 0
	couponCode := ""
	if req.CouponCode != "" {
		validateResp, err := s.couponSvc.Validate(ctx, model.ValidateCouponRequest{
			Code:            req.CouponCode,
			OrderTotalCents: subtotalCents,
		})
		if err != nil {
			return nil, err
		}
		if !validateResp.Valid {
			return nil, fmt.Errorf("coupon error: %s", validateResp.Message)
		}
		discountCents = validateResp.DiscountCents
		couponCode = req.CouponCode
	}

	totalCents := subtotalCents - discountCents
	if totalCents < 0 {
		totalCents = 0
	}

	paymentStatus := model.PaymentStatusPending
	if paymentMethod == model.PaymentMethodCOD {
		paymentStatus = model.PaymentStatusPending
	}

	order := &model.Order{
		UserID:            userID,
		TotalCents:        totalCents,
		DiscountCents:     discountCents,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     paymentStatus,
		CouponCode:        couponCode,
		ShippingName:      req.ShippingName,
		ShippingAddress:   req.ShippingAddress,
		CustomizationNote: req.CustomizationNote,
	}

	createdOrder, err := func() (*model.Order, error) {
		o := &model.Order{}
		return o, tx.QueryRow(ctx, `
			INSERT INTO orders (user_id, total_cents, discount_cents, payment_method, payment_status,
			                    coupon_code, shipping_name, shipping_address, customization_note)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING id, user_id, total_cents, discount_cents, status,
			          payment_method, payment_status, razorpay_order_id, coupon_code,
			          shipping_name, shipping_address, customization_note, created_at`,
			order.UserID, order.TotalCents, order.DiscountCents, order.PaymentMethod,
			order.PaymentStatus, order.CouponCode, order.ShippingName, order.ShippingAddress,
			order.CustomizationNote,
		).Scan(&o.ID, &o.UserID, &o.TotalCents, &o.DiscountCents, &o.Status,
			&o.PaymentMethod, &o.PaymentStatus, &o.RazorpayOrderID, &o.CouponCode,
			&o.ShippingName, &o.ShippingAddress, &o.CustomizationNote, &o.CreatedAt)
	}()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		var variantID *int64
		priceCents := 0
		if item.Variant != nil {
			variantID = &item.Variant.ID
			priceCents = item.Variant.PriceCents
		} else if item.Product != nil {
			priceCents = item.Product.PriceCents
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, variant_id, quantity, price_cents)
			VALUES ($1, $2, $3, $4, $5)`,
			createdOrder.ID, item.ProductID, variantID, item.Quantity, priceCents)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	_ = s.cartRepo.Clear(ctx, req.SessionID)

	// Fire-and-forget notifications
	if s.notif != nil && s.userRepo != nil {
		go func() {
			u, err := s.userRepo.FindByID(context.Background(), userID)
			if err == nil && u != nil {
				s.notif.SendOrderConfirmation(context.Background(), createdOrder, u.Identifier)
			}
			s.notif.SendVendorNewOrder(context.Background(), createdOrder)
		}()
	}

	return createdOrder, nil
}

func (s *OrderService) ListByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID)
}

func (s *OrderService) ListAll(ctx context.Context) ([]model.Order, error) {
	return s.orderRepo.ListAll(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	switch status {
	case model.OrderStatusPlaced, model.OrderStatusShipped, model.OrderStatusDelivered, model.OrderStatusCancelled:
	default:
		return fmt.Errorf("invalid status: %s", status)
	}
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		return err
	}

	// Fire-and-forget status notification
	if s.notif != nil && s.userRepo != nil {
		go func() {
			order, err := s.orderRepo.FindByID(context.Background(), orderID)
			if err != nil || order == nil {
				return
			}
			u, err := s.userRepo.FindByID(context.Background(), order.UserID)
			if err == nil && u != nil {
				s.notif.SendOrderStatusUpdate(context.Background(), order, u.Identifier, status)
			}
		}()
	}
	return nil
}
