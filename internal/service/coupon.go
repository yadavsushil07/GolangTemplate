package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type CouponService struct {
	repo *repository.CouponRepository
}

func NewCouponService(repo *repository.CouponRepository) *CouponService {
	return &CouponService{repo: repo}
}

func (s *CouponService) Validate(ctx context.Context, req model.ValidateCouponRequest) (*model.ValidateCouponResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return &model.ValidateCouponResponse{Valid: false, Message: "coupon code is required"}, nil
	}

	coupon, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return &model.ValidateCouponResponse{Valid: false, Message: "coupon not found"}, nil
	}
	if !coupon.IsActive {
		return &model.ValidateCouponResponse{Valid: false, Message: "coupon is no longer active"}, nil
	}
	if coupon.ExpiresAt != nil && time.Now().After(*coupon.ExpiresAt) {
		return &model.ValidateCouponResponse{Valid: false, Message: "coupon has expired"}, nil
	}
	if req.OrderTotalCents < coupon.MinOrderCents {
		return &model.ValidateCouponResponse{
			Valid:   false,
			Message: fmt.Sprintf("minimum order amount of ₹%d required", coupon.MinOrderCents/100),
		}, nil
	}

	discountCents := 0
	if coupon.DiscountPct != nil {
		discountCents = req.OrderTotalCents * (*coupon.DiscountPct) / 100
	} else if coupon.DiscountCents != nil {
		discountCents = *coupon.DiscountCents
		if discountCents > req.OrderTotalCents {
			discountCents = req.OrderTotalCents
		}
	}

	return &model.ValidateCouponResponse{
		Valid:         true,
		DiscountCents: discountCents,
		Message:       fmt.Sprintf("Coupon applied! You save ₹%d", discountCents/100),
	}, nil
}

func (s *CouponService) Create(ctx context.Context, req model.CreateCouponRequest) (*model.Coupon, error) {
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	if req.Code == "" {
		return nil, fmt.Errorf("coupon code is required")
	}
	if req.DiscountPct == nil && req.DiscountCents == nil {
		return nil, fmt.Errorf("either discount_pct or discount_cents is required")
	}
	if req.DiscountPct != nil && req.DiscountCents != nil {
		return nil, fmt.Errorf("cannot set both discount_pct and discount_cents")
	}
	if req.DiscountPct != nil && (*req.DiscountPct <= 0 || *req.DiscountPct > 100) {
		return nil, fmt.Errorf("discount_pct must be between 1 and 100")
	}
	return s.repo.Create(ctx, req)
}

func (s *CouponService) List(ctx context.Context) ([]model.Coupon, error) {
	return s.repo.List(ctx)
}

func (s *CouponService) Deactivate(ctx context.Context, id int64) error {
	return s.repo.Deactivate(ctx, id)
}
