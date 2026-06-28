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

type AdminHandler struct {
	userRepo  *repository.UserRepository
	orderRepo *repository.OrderRepository
	orderSvc  *service.OrderService
}

func NewAdminHandler(userRepo *repository.UserRepository, orderRepo *repository.OrderRepository, orderSvc *service.OrderService) *AdminHandler {
	return &AdminHandler{userRepo: userRepo, orderRepo: orderRepo, orderSvc: orderSvc}
}

// GET /api/admin/users?role=vendor&limit=50&offset=0
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	roleFilter := r.URL.Query().Get("role")
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	users, err := h.userRepo.List(r.Context(), roleFilter, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	if users == nil {
		users = []model.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

// PUT /api/admin/users/:id/role
func (h *AdminHandler) SetUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req model.SetRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch req.Role {
	case model.RoleCustomer, model.RoleVendor, model.RoleAdmin:
	default:
		writeError(w, http.StatusBadRequest, "invalid role: must be customer, vendor, or admin")
		return
	}

	if err := h.userRepo.SetRole(r.Context(), id, req.Role); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "role updated to " + req.Role,
	})
}

// GET /api/admin/summary
func (h *AdminHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	totalUsers, err := h.userRepo.Count(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count users")
		return
	}

	orders, err := h.orderRepo.ListAll(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	totalOrders := len(orders)
	totalRevenue := 0
	pending := 0
	shipped := 0
	for _, o := range orders {
		totalRevenue += o.TotalCents
		switch o.Status {
		case model.OrderStatusPlaced:
			pending++
		case model.OrderStatusShipped:
			shipped++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total_users":         totalUsers,
		"total_orders":        totalOrders,
		"total_revenue_cents": totalRevenue,
		"pending_orders":      pending,
		"shipped_orders":      shipped,
	})
}
