package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yadavsushil07/GolangTemplate/internal/middleware"
	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
)

type OrderHandler struct {
	orderSvc   *service.OrderService
	couponSvc  *service.CouponService
	paymentSvc *service.PaymentService
}

func NewOrderHandler(orderSvc *service.OrderService, couponSvc *service.CouponService, paymentSvc *service.PaymentService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc, couponSvc: couponSvc, paymentSvc: paymentSvc}
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SessionID == "" {
		req.SessionID = sessionIDFromRequest(r)
	}

	order, err := h.orderSvc.Checkout(r.Context(), userID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, err := h.orderSvc.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req model.ValidateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.couponSvc.Validate(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to validate coupon")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *OrderHandler) CreateRazorpayOrder(w http.ResponseWriter, r *http.Request) {
	if h.paymentSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "online payment not configured")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	_ = userID

	var body struct {
		OrderID     int64 `json:"order_id"`
		AmountCents int   `json:"amount_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rzpOrderID, keyID, err := h.paymentSvc.CreateRazorpayOrder(r.Context(), body.OrderID, body.AmountCents)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment order")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"razorpay_order_id": rzpOrderID,
		"key_id":            keyID,
	})
}

func (h *OrderHandler) VerifyRazorpayPayment(w http.ResponseWriter, r *http.Request) {
	if h.paymentSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "online payment not configured")
		return
	}
	var req model.RazorpayVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.paymentSvc.VerifyPayment(r.Context(), req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "payment verified"})
}
