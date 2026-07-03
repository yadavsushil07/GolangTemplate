package service

import (
	"fmt"
	"strings"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
)

// baseStyle returns common CSS used across all templates.
const baseStyle = `
  body { margin:0; padding:0; background:#f0f4f8; font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif; }
  .wrap { max-width:600px; margin:32px auto; background:#fff; border-radius:12px; overflow:hidden; box-shadow:0 4px 24px rgba(0,0,0,.08); }
  .header { background:#0f172a; padding:28px 32px; }
  .header h1 { color:#38bdf8; margin:0; font-size:22px; letter-spacing:.5px; }
  .header p { color:#94a3b8; margin:4px 0 0; font-size:13px; }
  .body { padding:32px; }
  .badge { display:inline-block; background:#38bdf822; color:#0284c7; border:1px solid #38bdf844; border-radius:999px; font-size:12px; font-weight:700; padding:4px 14px; letter-spacing:.5px; margin-bottom:20px; }
  h2 { color:#0f172a; font-size:20px; margin:0 0 8px; }
  p { color:#475569; font-size:14px; line-height:1.6; margin:0 0 16px; }
  .table { width:100%%; border-collapse:collapse; margin:20px 0; }
  .table th { background:#f8fafc; color:#64748b; font-size:12px; text-transform:uppercase; letter-spacing:.5px; padding:10px 12px; text-align:left; border-bottom:1px solid #e2e8f0; }
  .table td { padding:12px; border-bottom:1px solid #f1f5f9; color:#1e293b; font-size:14px; }
  .total-row td { font-weight:700; font-size:15px; color:#0f172a; border-top:2px solid #e2e8f0; }
  .info-box { background:#f8fafc; border:1px solid #e2e8f0; border-radius:8px; padding:16px; margin:16px 0; }
  .info-box p { margin:4px 0; font-size:13px; color:#475569; }
  .info-box strong { color:#1e293b; }
  .status-chip { display:inline-block; padding:6px 16px; border-radius:999px; font-size:13px; font-weight:700; }
  .status-placed    { background:#dbeafe; color:#1d4ed8; }
  .status-shipped   { background:#fef9c3; color:#a16207; }
  .status-delivered { background:#dcfce7; color:#15803d; }
  .status-cancelled { background:#fee2e2; color:#b91c1c; }
  .footer { background:#f8fafc; padding:20px 32px; text-align:center; color:#94a3b8; font-size:12px; border-top:1px solid #e2e8f0; }
  .btn { display:inline-block; background:#0284c7; color:#fff; text-decoration:none; padding:12px 28px; border-radius:8px; font-weight:700; font-size:14px; margin-top:8px; }
`

func htmlPage(content string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"/><style>%s</style></head><body><div class="wrap">%s<div class="footer">SBY TWILIGHT &copy; 2026 &nbsp;|&nbsp; Kurla West, Mumbai 400070</div></div></body></html>`, baseStyle, content)
}

// otpEmailHTML generates the OTP email body.
func otpEmailHTML(otp string) string {
	content := fmt.Sprintf(`
<div class="header"><h1>SBY TWILIGHT</h1><p>Your login code</p></div>
<div class="body">
  <div class="badge">OTP</div>
  <h2>Your One-Time Password</h2>
  <p>Use the code below to log in to your SBY TWILIGHT account. It is valid for <strong>5 minutes</strong>.</p>
  <div style="text-align:center;margin:28px 0;">
    <div style="font-size:42px;font-weight:900;letter-spacing:12px;color:#0f172a;background:#f8fafc;border:2px dashed #e2e8f0;border-radius:12px;padding:20px;">%s</div>
  </div>
  <p style="font-size:13px;color:#94a3b8;">If you did not request this, please ignore this email.</p>
</div>`, otp)
	return htmlPage(content)
}

// orderConfirmationHTML generates the order confirmation email.
func orderConfirmationHTML(order *model.Order) string {
	rows := ""
	for _, item := range order.Items {
		name := fmt.Sprintf("Product #%d", item.ProductID)
		if item.Product != nil {
			name = item.Product.Name
		}
		variant := ""
		if item.Variant != nil {
			variant = fmt.Sprintf(" (%s", item.Variant.Size)
			if item.Variant.Color != "" {
				variant += " · " + item.Variant.Color
			}
			variant += ")"
		}
		rows += fmt.Sprintf(`<tr><td>%s%s</td><td style="text-align:center">%d</td><td style="text-align:right">₹%d</td></tr>`,
			name, variant, item.Quantity, item.PriceCents*item.Quantity/100)
	}

	discountRow := ""
	if order.DiscountCents > 0 {
		discountRow = fmt.Sprintf(`<tr class="total-row"><td colspan="2">Discount (%s)</td><td style="text-align:right;color:#15803d">-₹%d</td></tr>`,
			order.CouponCode, order.DiscountCents/100)
	}

	paymentBadge := "Cash on Delivery"
	if order.PaymentMethod == model.PaymentMethodRazorpay {
		paymentBadge = "Online Payment (Razorpay)"
	}

	note := ""
	if order.CustomizationNote != "" {
		note = fmt.Sprintf(`<div class="info-box"><p><strong>Customization Note:</strong> %s</p></div>`, order.CustomizationNote)
	}

	content := fmt.Sprintf(`
<div class="header"><h1>SBY TWILIGHT</h1><p>Order Confirmation</p></div>
<div class="body">
  <div class="badge">ORDER PLACED</div>
  <h2>Thank you for your order!</h2>
  <p>Your order <strong>#%d</strong> has been confirmed and is being prepared.</p>
  <table class="table">
    <thead><tr><th>Item</th><th style="text-align:center">Qty</th><th style="text-align:right">Price</th></tr></thead>
    <tbody>%s%s<tr class="total-row"><td colspan="2"><strong>Total</strong></td><td style="text-align:right">₹%d</td></tr></tbody>
  </table>
  <div class="info-box">
    <p><strong>Ship to:</strong> %s</p>
    <p>%s</p>
    <p><strong>Payment:</strong> %s</p>
  </div>
  %s
  <p style="font-size:13px;color:#94a3b8;">We will send you another update when your order ships.</p>
</div>`,
		order.ID, rows, discountRow, order.TotalCents/100,
		order.ShippingName, order.ShippingAddress, paymentBadge, note)
	return htmlPage(content)
}

// orderStatusHTML generates the status update email.
func orderStatusHTML(order *model.Order, status string) string {
	titles := map[string]string{
		model.OrderStatusShipped:   "Your Order Has Shipped!",
		model.OrderStatusDelivered: "Your Order Has Been Delivered",
		model.OrderStatusCancelled: "Your Order Has Been Cancelled",
	}
	messages := map[string]string{
		model.OrderStatusShipped:   fmt.Sprintf("Great news! Your order <strong>#%d</strong> is on its way. Estimated delivery in 3–7 business days.", order.ID),
		model.OrderStatusDelivered: fmt.Sprintf("Your order <strong>#%d</strong> has been delivered. We hope you love it! 🙏", order.ID),
		model.OrderStatusCancelled: fmt.Sprintf("Your order <strong>#%d</strong> has been cancelled. If you paid online, a refund will be processed in 5–7 business days.", order.ID),
	}
	title := titles[status]
	msg := messages[status]

	steps := []struct{ label, st string }{
		{"Order Placed", model.OrderStatusPlaced},
		{"Shipped", model.OrderStatusShipped},
		{"Delivered", model.OrderStatusDelivered},
	}
	timeline := `<div style="display:flex;gap:8px;margin:24px 0;flex-wrap:wrap;">`
	for _, step := range steps {
		chipClass := "status-chip"
		if step.st == status {
			chipClass += " status-" + step.st
		} else {
			chipClass = "status-chip"
		}
		active := ""
		if step.st == status {
			active = fmt.Sprintf(` class="%s status-%s"`, "status-chip", step.st)
		} else {
			active = ` style="background:#f1f5f9;color:#94a3b8;display:inline-block;padding:6px 16px;border-radius:999px;font-size:13px;font-weight:700;"`
		}
		_ = chipClass
		timeline += fmt.Sprintf(`<span%s>%s</span>`, active, step.label)
	}
	timeline += `</div>`

	content := fmt.Sprintf(`
<div class="header"><h1>SBY TWILIGHT</h1><p>Order Update</p></div>
<div class="body">
  <div class="badge">%s</div>
  <h2>%s</h2>
  <p>%s</p>
  %s
  <div class="info-box">
    <p><strong>Order #%d</strong></p>
    <p><strong>Ship to:</strong> %s, %s</p>
  </div>
</div>`,
		strings.ToUpper(status), title, msg, timeline,
		order.ID, order.ShippingName, order.ShippingAddress)
	return htmlPage(content)
}

// vendorAlertHTML generates the new-order alert for the vendor.
func vendorAlertHTML(order *model.Order) string {
	rows := ""
	for _, item := range order.Items {
		name := fmt.Sprintf("Product #%d", item.ProductID)
		if item.Product != nil {
			name = item.Product.Name
		}
		variant := ""
		if item.Variant != nil {
			variant = fmt.Sprintf(" (%s", item.Variant.Size)
			if item.Variant.Color != "" {
				variant += " · " + item.Variant.Color
			}
			variant += ")"
		}
		rows += fmt.Sprintf(`<tr><td>%s%s</td><td style="text-align:center">%d</td><td style="text-align:right">₹%d</td></tr>`,
			name, variant, item.Quantity, item.PriceCents*item.Quantity/100)
	}

	note := ""
	if order.CustomizationNote != "" {
		note = fmt.Sprintf(`<div class="info-box" style="border-color:#fbbf24;background:#fffbeb;"><p><strong>⚠ Customization Note:</strong> %s</p></div>`, order.CustomizationNote)
	}

	content := fmt.Sprintf(`
<div class="header"><h1>SBY TWILIGHT</h1><p>New Order Alert</p></div>
<div class="body">
  <div class="badge">NEW ORDER</div>
  <h2>Order #%d Received</h2>
  <p>A new order has been placed. Please prepare and ship it as soon as possible.</p>
  <table class="table">
    <thead><tr><th>Item</th><th style="text-align:center">Qty</th><th style="text-align:right">Price</th></tr></thead>
    <tbody>%s<tr class="total-row"><td colspan="2"><strong>Total</strong></td><td style="text-align:right">₹%d</td></tr></tbody>
  </table>
  <div class="info-box">
    <p><strong>Ship to:</strong> %s</p>
    <p>%s</p>
    <p><strong>Payment:</strong> %s</p>
  </div>
  %s
</div>`,
		order.ID, rows, order.TotalCents/100,
		order.ShippingName, order.ShippingAddress,
		func() string {
			if order.PaymentMethod == model.PaymentMethodRazorpay {
				return "Online (Razorpay) — " + order.PaymentStatus
			}
			return "Cash on Delivery"
		}(), note)
	return htmlPage(content)
}
