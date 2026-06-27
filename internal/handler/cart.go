package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yadavsushil07/GolangTemplate/internal/service"
)

type CartHandler struct {
	cartSvc *service.CartService
}

func NewCartHandler(cartSvc *service.CartService) *CartHandler {
	return &CartHandler{cartSvc: cartSvc}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	sessionID := sessionIDFromRequest(r)
	summary, err := h.cartSvc.GetCart(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get cart")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sessionID := sessionIDFromRequest(r)
	if err := h.cartSvc.AddItem(r.Context(), sessionID, body.ProductID, body.Quantity); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	summary, err := h.cartSvc.GetCart(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get cart")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "productId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	sessionID := sessionIDFromRequest(r)
	if err := h.cartSvc.RemoveItem(r.Context(), sessionID, productID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove item")
		return
	}

	summary, err := h.cartSvc.GetCart(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get cart")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func sessionIDFromRequest(r *http.Request) string {
	if c, err := r.Cookie("session_id"); err == nil && c.Value != "" {
		return c.Value
	}
	return r.Header.Get("X-Session-ID")
}
