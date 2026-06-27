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
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{cartRepo: cartRepo, productRepo: productRepo}
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
		if item.Product != nil {
			total += item.Product.PriceCents * item.Quantity
		}
	}
	return &model.CartSummary{Items: items, TotalCents: total}, nil
}

func (s *CartService) AddItem(ctx context.Context, sessionID string, productID int64, qty int) error {
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
	return s.cartRepo.Upsert(ctx, sessionID, productID, qty)
}

func (s *CartService) RemoveItem(ctx context.Context, sessionID string, productID int64) error {
	return s.cartRepo.Remove(ctx, sessionID, productID)
}

func (s *CartService) Clear(ctx context.Context, sessionID string) error {
	return s.cartRepo.Clear(ctx, sessionID)
}
