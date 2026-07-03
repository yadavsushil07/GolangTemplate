package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

// NotificationService sends OTP and transactional messages via Fast2SMS and Resend.
type NotificationService struct {
	fast2smsKey  string
	resendKey    string
	fromEmail    string
	fromName     string
	vendorEmail  string
	vendorPhone  string
	httpClient   *http.Client
}

type NotificationConfig struct {
	Fast2SMSKey string
	ResendKey   string
	FromEmail   string
	FromName    string
	VendorEmail string
	VendorPhone string
}

func NewNotificationService(cfg NotificationConfig) *NotificationService {
	return &NotificationService{
		fast2smsKey: cfg.Fast2SMSKey,
		resendKey:   cfg.ResendKey,
		fromEmail:   cfg.FromEmail,
		fromName:    cfg.FromName,
		vendorEmail: cfg.VendorEmail,
		vendorPhone: cfg.VendorPhone,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// IsPhone returns true if the identifier looks like a phone number.
func IsPhone(identifier string) bool {
	trimmed := strings.TrimPrefix(identifier, "+")
	if len(trimmed) < 10 || len(trimmed) > 13 {
		return false
	}
	for _, c := range trimmed {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// SendOTP delivers the OTP via SMS (phone) or email — detected from identifier format.
func (n *NotificationService) SendOTP(identifier, otp string) {
	if IsPhone(identifier) {
		if n.fast2smsKey == "" {
			log.Printf("[OTP] %s → %s (Fast2SMS not configured)", identifier, otp)
			return
		}
		go func() {
			if err := n.sendSMS(identifier, fmt.Sprintf("Your SBY TWILIGHT OTP is %s. Valid for 5 minutes.", otp)); err != nil {
				log.Printf("[NotificationService] SMS OTP failed for %s: %v", identifier, err)
			}
		}()
	} else {
		if n.resendKey == "" {
			log.Printf("[OTP] %s → %s (Resend not configured)", identifier, otp)
			return
		}
		go func() {
			subject := "Your SBY TWILIGHT OTP"
			body := otpEmailHTML(otp)
			if err := n.sendEmail(identifier, subject, body); err != nil {
				log.Printf("[NotificationService] Email OTP failed for %s: %v", identifier, err)
			}
		}()
	}
}

// SendOrderConfirmation notifies the customer after a successful order.
func (n *NotificationService) SendOrderConfirmation(ctx context.Context, order *model.Order, identifier string) {
	msg := fmt.Sprintf("Order #%d confirmed! Total: ₹%d. We will update you when it ships. - SBY TWILIGHT",
		order.ID, order.TotalCents/100)

	if IsPhone(identifier) {
		go func() {
			if err := n.sendSMS(identifier, msg); err != nil {
				log.Printf("[NotificationService] order confirmation SMS failed: %v", err)
			}
		}()
	} else {
		go func() {
			if err := n.sendEmail(identifier, fmt.Sprintf("Order #%d Confirmed – SBY TWILIGHT", order.ID), orderConfirmationHTML(order)); err != nil {
				log.Printf("[NotificationService] order confirmation email failed: %v", err)
			}
		}()
	}
}

// SendOrderStatusUpdate notifies the customer when order status changes.
func (n *NotificationService) SendOrderStatusUpdate(ctx context.Context, order *model.Order, identifier, status string) {
	statusMsg := map[string]string{
		model.OrderStatusShipped:   fmt.Sprintf("Order #%d has been shipped! You'll receive it soon. - SBY TWILIGHT", order.ID),
		model.OrderStatusDelivered: fmt.Sprintf("Order #%d delivered! Thank you for shopping with SBY TWILIGHT. 🙏", order.ID),
		model.OrderStatusCancelled: fmt.Sprintf("Order #%d has been cancelled. Refund (if any) will be processed in 5-7 days. - SBY TWILIGHT", order.ID),
	}
	msg, ok := statusMsg[status]
	if !ok {
		return
	}

	if IsPhone(identifier) {
		go func() {
			if err := n.sendSMS(identifier, msg); err != nil {
				log.Printf("[NotificationService] status SMS failed: %v", err)
			}
		}()
	} else {
		go func() {
			subject := fmt.Sprintf("Order #%d %s – SBY TWILIGHT", order.ID, strings.Title(status))
			if err := n.sendEmail(identifier, subject, orderStatusHTML(order, status)); err != nil {
				log.Printf("[NotificationService] status email failed: %v", err)
			}
		}()
	}
}

// SendVendorNewOrder notifies the vendor when a new order is placed.
func (n *NotificationService) SendVendorNewOrder(ctx context.Context, order *model.Order) {
	if n.vendorPhone != "" {
		go func() {
			msg := fmt.Sprintf("New order #%d received! Total: ₹%d. Ship to: %s. - SBY TWILIGHT",
				order.ID, order.TotalCents/100, order.ShippingName)
			if err := n.sendSMS(n.vendorPhone, msg); err != nil {
				log.Printf("[NotificationService] vendor SMS failed: %v", err)
			}
		}()
	}
	if n.vendorEmail != "" {
		go func() {
			subject := fmt.Sprintf("New Order #%d – SBY TWILIGHT", order.ID)
			if err := n.sendEmail(n.vendorEmail, subject, vendorAlertHTML(order)); err != nil {
				log.Printf("[NotificationService] vendor email failed: %v", err)
			}
		}()
	}
}

// ---- Internal transports ----

func (n *NotificationService) sendSMS(phone, message string) error {
	if n.fast2smsKey == "" {
		return fmt.Errorf("Fast2SMS key not configured")
	}
	// Fast2SMS Quick SMS API
	payload := map[string]string{
		"route":   "q",
		"message": message,
		"numbers": phone,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://www.fast2sms.com/dev/bulkV2", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", n.fast2smsKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fast2sms error %d: %s", resp.StatusCode, respBody)
	}
	log.Printf("[NotificationService] SMS sent to %s", phone)
	return nil
}

func (n *NotificationService) sendEmail(to, subject, htmlBody string) error {
	if n.resendKey == "" {
		return fmt.Errorf("Resend key not configured")
	}
	fromField := n.fromEmail
	if n.fromName != "" {
		fromField = fmt.Sprintf("%s <%s>", n.fromName, n.fromEmail)
	}
	payload := map[string]any{
		"from":    fromField,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+n.resendKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend error %d: %s", resp.StatusCode, respBody)
	}
	log.Printf("[NotificationService] Email sent to %s (subject: %s)", to, subject)
	return nil
}
