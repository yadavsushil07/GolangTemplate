package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	razorpay "github.com/razorpay/razorpay-go"
)

type PaymentService struct {
	client    *razorpay.Client
	keySecret string
	orderRepo *repository.OrderRepository
}

func NewPaymentService(keyID, keySecret string, orderRepo *repository.OrderRepository) *PaymentService {
	return &PaymentService{
		client:    razorpay.NewClient(keyID, keySecret),
		keySecret: keySecret,
		orderRepo: orderRepo,
	}
}

// CreateRazorpayOrder creates a Razorpay order and stores the razorpay_order_id on the local order.
func (s *PaymentService) CreateRazorpayOrder(ctx context.Context, orderID int64, amountCents int) (string, string, error) {
	data := map[string]any{
		"amount":   amountCents,
		"currency": "INR",
		"receipt":  fmt.Sprintf("order_%d", orderID),
	}
	body, err := s.client.Order.Create(data, nil)
	if err != nil {
		return "", "", fmt.Errorf("razorpay order creation failed: %w", err)
	}

	rzpID, ok := body["id"].(string)
	if !ok || rzpID == "" {
		return "", "", fmt.Errorf("invalid razorpay response")
	}

	if err := s.orderRepo.UpdateRazorpayOrderID(ctx, orderID, rzpID); err != nil {
		return "", "", err
	}

	return rzpID, strings.Split(s.keySecret, "|")[0], nil
}

// VerifyPayment verifies the Razorpay payment signature and marks the order as paid.
func (s *PaymentService) VerifyPayment(ctx context.Context, req model.RazorpayVerifyRequest) error {
	expected := req.RazorpayOrderID + "|" + req.PaymentID
	h := hmac.New(sha256.New, []byte(s.keySecret))
	h.Write([]byte(expected))
	computed := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(computed), []byte(req.Signature)) {
		return fmt.Errorf("payment signature verification failed")
	}

	order, err := s.orderRepo.FindByRazorpayOrderID(ctx, req.RazorpayOrderID)
	if err != nil || order == nil {
		return fmt.Errorf("order not found for razorpay_order_id %s", req.RazorpayOrderID)
	}

	return s.orderRepo.UpdatePaymentStatus(ctx, order.ID, model.PaymentStatusPaid)
}
