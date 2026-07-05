package service

import (
	"context"
	"fmt"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type ProductService struct {
	repo        *repository.ProductRepository
	variantRepo *repository.VariantRepository
	catRepo     *repository.CategoryRepository
}

func NewProductService(repo *repository.ProductRepository, variantRepo *repository.VariantRepository, catRepo *repository.CategoryRepository) *ProductService {
	return &ProductService{repo: repo, variantRepo: variantRepo, catRepo: catRepo}
}

func (s *ProductService) List(ctx context.Context, activeOnly bool, categorySlug string) ([]model.Product, error) {
	return s.repo.List(ctx, activeOnly, categorySlug)
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

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*model.Product, error) {
	p, err := s.repo.FindBySlug(ctx, slug)
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
	for i, v := range req.Variants {
		if v.PriceCents <= 0 {
			return nil, fmt.Errorf("variant %d: price must be greater than zero", i+1)
		}
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

func (s *ProductService) AddVariant(ctx context.Context, productID int64, req model.CreateVariantRequest) (*model.ProductVariant, error) {
	if req.Size == "" {
		return nil, fmt.Errorf("size is required")
	}
	if req.PriceCents <= 0 {
		return nil, fmt.Errorf("price must be greater than zero")
	}
	return s.variantRepo.Create(ctx, productID, req)
}

func (s *ProductService) DeleteVariant(ctx context.Context, variantID int64) error {
	return s.variantRepo.Delete(ctx, variantID)
}

func (s *ProductService) AddImages(ctx context.Context, productID int64, urls []string) error {
	return s.variantRepo.AddImages(ctx, productID, urls)
}

func (s *ProductService) DeleteImage(ctx context.Context, productID, imageID int64) error {
	return s.variantRepo.DeleteImage(ctx, productID, imageID)
}

func (s *ProductService) SetCategories(ctx context.Context, productID int64, categoryIDs []int64) error {
	return s.catRepo.SetProductCategories(ctx, productID, categoryIDs)
}
