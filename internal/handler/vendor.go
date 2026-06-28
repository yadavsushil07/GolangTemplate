package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
)

type VendorHandler struct {
	productSvc *service.ProductService
	orderSvc   *service.OrderService
	couponSvc  *service.CouponService
	catRepo    *repository.CategoryRepository
}

func NewVendorHandler(productSvc *service.ProductService, orderSvc *service.OrderService, couponSvc *service.CouponService, catRepo *repository.CategoryRepository) *VendorHandler {
	return &VendorHandler{productSvc: productSvc, orderSvc: orderSvc, couponSvc: couponSvc, catRepo: catRepo}
}

// ---- Products ----

func (h *VendorHandler) ListAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productSvc.List(r.Context(), false, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch products")
		return
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *VendorHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	product, err := h.productSvc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (h *VendorHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	product, err := h.productSvc.Update(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *VendorHandler) DeactivateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := h.productSvc.Deactivate(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to deactivate product")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product deactivated"})
}

// ---- Variants ----

func (h *VendorHandler) AddVariant(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var req model.CreateVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	v, err := h.productSvc.AddVariant(r.Context(), productID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

func (h *VendorHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	variantID, err := strconv.ParseInt(chi.URLParam(r, "variantId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid variant id")
		return
	}
	if err := h.productSvc.DeleteVariant(r.Context(), variantID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete variant")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "variant deleted"})
}

// ---- Images ----

func (h *VendorHandler) AddImages(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var body struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.productSvc.AddImages(r.Context(), productID, body.URLs); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add images")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "images added"})
}

func (h *VendorHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	imageID, err := strconv.ParseInt(chi.URLParam(r, "imageId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid image id")
		return
	}
	if err := h.productSvc.DeleteImage(r.Context(), imageID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete image")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "image deleted"})
}

// ---- Categories ----

func (h *VendorHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.catRepo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *VendorHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c, err := h.catRepo.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *VendorHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}
	if err := h.catRepo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete category")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "category deleted"})
}

func (h *VendorHandler) SetProductCategories(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var body struct {
		CategoryIDs []int64 `json:"category_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.productSvc.SetCategories(r.Context(), productID, body.CategoryIDs); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set categories")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "categories updated"})
}

// ---- Orders ----

func (h *VendorHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderSvc.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *VendorHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.orderSvc.UpdateStatus(r.Context(), id, body.Status); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "order status updated"})
}

// ---- Coupons ----

func (h *VendorHandler) ListCoupons(w http.ResponseWriter, r *http.Request) {
	coupons, err := h.couponSvc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch coupons")
		return
	}
	writeJSON(w, http.StatusOK, coupons)
}

func (h *VendorHandler) CreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c, err := h.couponSvc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *VendorHandler) DeactivateCoupon(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid coupon id")
		return
	}
	if err := h.couponSvc.Deactivate(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to deactivate coupon")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "coupon deactivated"})
}
