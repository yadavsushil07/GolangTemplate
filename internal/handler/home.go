package handler

import (
	"fmt"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, homePage)
}

const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>AaryaShop API</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #0f172a;
      color: #f8fafc;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 24px;
    }
    .card {
      background: #1e293b;
      border: 1px solid #334155;
      border-radius: 16px;
      padding: 40px 48px;
      max-width: 640px;
      width: 100%;
      box-shadow: 0 25px 50px rgba(0,0,0,0.4);
    }
    .badge {
      display: inline-block;
      background: #38bdf822;
      color: #38bdf8;
      border: 1px solid #38bdf844;
      border-radius: 999px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 1px;
      text-transform: uppercase;
      padding: 4px 12px;
      margin-bottom: 20px;
    }
    h1 { font-size: 32px; font-weight: 800; margin-bottom: 8px; }
    .subtitle { color: #94a3b8; font-size: 15px; margin-bottom: 32px; line-height: 1.6; }
    .status {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #4ade80;
      font-size: 14px;
      font-weight: 600;
      margin-bottom: 32px;
    }
    .dot { width: 8px; height: 8px; border-radius: 50%; background: #4ade80; }
    h2 { font-size: 13px; font-weight: 700; color: #94a3b8; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 12px; }
    .routes { list-style: none; display: flex; flex-direction: column; gap: 8px; margin-bottom: 28px; }
    .routes li {
      display: flex;
      align-items: center;
      gap: 10px;
      background: #0f172a;
      border: 1px solid #334155;
      border-radius: 8px;
      padding: 10px 14px;
      font-size: 13px;
      font-family: "SF Mono", "Fira Code", monospace;
    }
    .method {
      font-size: 11px;
      font-weight: 800;
      border-radius: 4px;
      padding: 2px 7px;
      text-transform: uppercase;
    }
    .get  { background: #14532d; color: #4ade80; }
    .post { background: #1e3a5f; color: #60a5fa; }
    .put  { background: #4a1d96; color: #c4b5fd; }
    .del  { background: #7f1d1d; color: #fca5a5; }
    .path { color: #e2e8f0; flex: 1; }
    .desc { color: #64748b; font-size: 12px; font-family: inherit; }
    footer { margin-top: 28px; color: #475569; font-size: 12px; text-align: center; }
  </style>
</head>
<body>
  <div class="card">
    <div class="badge">REST API</div>
    <h1>AaryaShop</h1>
    <p class="subtitle">Backend API for the AaryaShop e-commerce platform — products, variants, cart, orders, coupons &amp; payments.</p>
    <div class="status"><div class="dot"></div>Server is running</div>

    <h2>Public Endpoints</h2>
    <ul class="routes">
      <li><span class="method get">GET</span><span class="path">/api/products</span><span class="desc">List products (?category=slug)</span></li>
      <li><span class="method get">GET</span><span class="path">/api/products/:id</span><span class="desc">Product detail with variants &amp; images</span></li>
      <li><span class="method get">GET</span><span class="path">/api/categories</span><span class="desc">List categories</span></li>
      <li><span class="method get">GET</span><span class="path">/api/cart</span><span class="desc">Get cart (X-Session-ID header)</span></li>
      <li><span class="method post">POST</span><span class="path">/api/cart</span><span class="desc">Add item to cart</span></li>
      <li><span class="method post">POST</span><span class="path">/api/auth/request-otp</span><span class="desc">Request OTP</span></li>
      <li><span class="method post">POST</span><span class="path">/api/auth/verify-otp</span><span class="desc">Verify OTP → JWT</span></li>
      <li><span class="method post">POST</span><span class="path">/api/coupons/validate</span><span class="desc">Validate coupon code</span></li>
    </ul>

    <h2>Customer (JWT required)</h2>
    <ul class="routes">
      <li><span class="method post">POST</span><span class="path">/api/checkout</span><span class="desc">Place order (COD or Razorpay)</span></li>
      <li><span class="method get">GET</span><span class="path">/api/orders</span><span class="desc">My order history</span></li>
      <li><span class="method post">POST</span><span class="path">/api/payments/razorpay/create-order</span><span class="desc">Init Razorpay order</span></li>
      <li><span class="method post">POST</span><span class="path">/api/payments/razorpay/verify</span><span class="desc">Verify payment signature</span></li>
    </ul>

    <h2>Vendor (JWT + vendor role)</h2>
    <ul class="routes">
      <li><span class="method post">POST</span><span class="path">/api/vendor/products</span><span class="desc">Create product</span></li>
      <li><span class="method put">PUT</span><span class="path">/api/vendor/products/:id</span><span class="desc">Update product</span></li>
      <li><span class="method post">POST</span><span class="path">/api/vendor/products/:id/variants</span><span class="desc">Add size/color variant</span></li>
      <li><span class="method post">POST</span><span class="path">/api/vendor/products/:id/images</span><span class="desc">Add product images</span></li>
      <li><span class="method post">POST</span><span class="path">/api/vendor/coupons</span><span class="desc">Create coupon</span></li>
      <li><span class="method get">GET</span><span class="path">/api/vendor/orders</span><span class="desc">All orders</span></li>
      <li><span class="method put">PUT</span><span class="path">/api/vendor/orders/:id/status</span><span class="desc">Update order status</span></li>
    </ul>

    <footer>AaryaShop &copy; 2026 &nbsp;|&nbsp; Built with Go + PostgreSQL</footer>
  </div>
</body>
</html>`
