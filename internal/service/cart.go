package service

import (
	"context"
	"fmt"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
	variantRepo *repository.VariantRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository, variantRepo *repository.VariantRepository) *CartService {
	return &CartService{cartRepo: cartRepo, productRepo: productRepo, variantRepo: variantRepo}
}

func (s *CartService) GetCart(ctx context.Context, sessionID string) (*model.CartSummary, error) {
	items, err := s.cartRepo.GetItems(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []model.CartItem{}
	}
	total := 0
	for _, item := range items {
		if item.Variant != nil {
			total += item.Variant.PriceCents * item.Quantity
		} else if item.Product != nil {
			total += item.Product.PriceCents * item.Quantity
		}
	}
	return &model.CartSummary{Items: items, TotalCents: total}, nil
}

func (s *CartService) AddItem(ctx context.Context, sessionID string, productID int64, variantID *int64, qty int) error {
	if qty <= 0 {
		qty = 1
	}
	p, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if p == nil || !p.IsActive {
		return fmt.Errorf("product not available")
	}

	if variantID != nil {
		v, err := s.variantRepo.FindByID(ctx, *variantID)
		if err != nil {
			return err
		}
		if v == nil || !v.IsActive {
			return fmt.Errorf("variant not available")
		}
		if v.Stock < qty {
			return fmt.Errorf("insufficient stock for this variant")
		}
	}

	return s.cartRepo.Upsert(ctx, sessionID, productID, variantID, qty)
}

func (s *CartService) RemoveItem(ctx context.Context, sessionID string, productID int64) error {
	return s.cartRepo.Remove(ctx, sessionID, productID)
}

func (s *CartService) RemoveVariant(ctx context.Context, sessionID string, variantID int64) error {
	return s.cartRepo.RemoveVariant(ctx, sessionID, variantID)
}

func (s *CartService) Clear(ctx context.Context, sessionID string) error {
	return s.cartRepo.Clear(ctx, sessionID)
}
