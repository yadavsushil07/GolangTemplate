package service

import (
	"context"
	"fmt"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context, activeOnly bool) ([]model.Product, error) {
	return s.repo.List(ctx, activeOnly)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("product not found")
	}
	return p, nil
}

func (s *ProductService) Create(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if req.PriceCents <= 0 {
		return nil, fmt.Errorf("price must be greater than zero")
	}
	return s.repo.Create(ctx, req)
}

func (s *ProductService) Update(ctx context.Context, id int64, req model.UpdateProductRequest) (*model.Product, error) {
	p, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("product not found")
	}
	return p, nil
}

func (s *ProductService) Deactivate(ctx context.Context, id int64) error {
	return s.repo.Deactivate(ctx, id)
}
