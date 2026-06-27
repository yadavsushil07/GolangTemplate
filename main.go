package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const jwtSecret = "dev-secret-change-me"

type product struct {
	ID          string
	Name        string
	Description string
	PriceCents  int
}

type otpEntry struct {
	Code      string
	ExpiresAt time.Time
}

type rateLimiter struct {
	window time.Duration
	limit  int
	mu     sync.Mutex
	items  map[string][]time.Time
}

type app struct {
	products []product
	carts    map[string]map[string]int
	otpStore map[string]otpEntry
	rl       *rateLimiter
}

func newRateLimiter(window time.Duration, limit int) *rateLimiter {
	return &rateLimiter{window: window, limit: limit, items: make(map[string][]time.Time)}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entries := rl.items[key]
	filtered := entries[:0]
	for _, entry := range entries {
		if now.Sub(entry) < rl.window {
			filtered = append(filtered, entry)
		}
	}
	filtered = append(filtered, now)
	rl.items[key] = filtered
	return len(filtered) <= rl.limit
}

func (a *app) init() {
	a.products = []product{
		{ID: "storm", Name: "Storm Jacket", Description: "Weatherproof outerwear for city nights", PriceCents: 8900},
		{ID: "aurora", Name: "Aurora Knit", Description: "Soft premium knit for weekend layering", PriceCents: 6400},
		{ID: "nova", Name: "Nova Sneakers", Description: "Lightweight street-ready comfort", PriceCents: 7200},
	}
	a.carts = make(map[string]map[string]int)
	a.otpStore = make(map[string]otpEntry)
	a.rl = newRateLimiter(1*time.Minute, 6)
}

func main() {
	a := &app{}
	a.init()

	mux := http.NewServeMux()
	mux.HandleFunc("/", a.homeHandler)
	mux.HandleFunc("/cart", a.cartHandler)
	mux.HandleFunc("/cart/add", a.addToCartHandler)
	mux.HandleFunc("/cart/remove", a.removeFromCartHandler)
	mux.HandleFunc("/login", a.loginHandler)
	mux.HandleFunc("/login/request-otp", a.requestOTPHandler)
	mux.HandleFunc("/login/verify-otp", a.verifyOTPHandler)
	mux.HandleFunc("/checkout", a.checkoutHandler)
	mux.HandleFunc("/logout", a.logoutHandler)

	handler := a.rateLimitMiddleware(a.authMiddleware(mux))
	log.Println("server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func (a *app) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	cartID := a.getCartID(w, r)
	cartCount := a.cartCount(cartID)
	user := a.currentUser(r)
	writePage(w, "Northstar Studio", buildHomePage(cartCount, user, a.products))
}

func (a *app) cartHandler(w http.ResponseWriter, r *http.Request) {
	cartID := a.getCartID(w, r)
	items := a.cartItems(cartID)
	user := a.currentUser(r)
	writePage(w, "Cart", buildCartPage(items, a.products, user, cartID))
}

func (a *app) addToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cartID := a.getCartID(w, r)
	productID := strings.TrimSpace(r.FormValue("product_id"))
	if productID == "" {
		http.Error(w, "missing product", http.StatusBadRequest)
		return
	}
	a.ensureCart(cartID)[productID]++
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (a *app) removeFromCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cartID := a.getCartID(w, r)
	productID := strings.TrimSpace(r.FormValue("product_id"))
	if productID != "" {
		cart := a.ensureCart(cartID)
		if cart[productID] > 1 {
			cart[productID]--
		} else {
			delete(cart, productID)
		}
	}
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (a *app) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writePage(w, "Login", buildLoginPage("", "", ""))
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (a *app) requestOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.rl.allow(getClientIP(r)) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	identifier := strings.TrimSpace(r.FormValue("identifier"))
	if identifier == "" {
		writePage(w, "Login", buildLoginPage("", "Please enter an email or phone number.", ""))
		return
	}
	code, err := a.requestOTP(identifier)
	if err != nil {
		writePage(w, "Login", buildLoginPage(identifier, err.Error(), ""))
		return
	}
	writePage(w, "Login", buildLoginPage(identifier, fmt.Sprintf("OTP ready: %s. Use it to finish sign in.", code), code))
}

func (a *app) verifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.rl.allow(getClientIP(r)) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	identifier := strings.TrimSpace(r.FormValue("identifier"))
	code := strings.TrimSpace(r.FormValue("code"))
	if identifier == "" || code == "" {
		writePage(w, "Login", buildLoginPage(identifier, "Please provide both the identifier and the OTP code.", ""))
		return
	}
	token, err := a.verifyOTP(identifier, code)
	if err != nil {
		writePage(w, "Login", buildLoginPage(identifier, err.Error(), ""))
		return
	}
	cookie := &http.Cookie{Name: "auth_token", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 60 * 60 * 24}
	http.SetCookie(w, cookie)
	next := r.FormValue("next")
	if next == "" {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (a *app) checkoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/checkout" {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodGet {
		cartID := a.getCartID(w, r)
		items := a.cartItems(cartID)
		if len(items) == 0 {
			writePage(w, "Checkout", buildCheckoutPage(items, a.products, "Your cart is empty.", ""))
			return
		}
		writePage(w, "Checkout", buildCheckoutPage(items, a.products, "", ""))
		return
	}
	if r.Method == http.MethodPost {
		cartID := a.getCartID(w, r)
		items := a.cartItems(cartID)
		if len(items) == 0 {
			writePage(w, "Checkout", buildCheckoutPage(items, a.products, "Your cart is empty.", ""))
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		card := strings.TrimSpace(r.FormValue("card"))
		if name == "" || card == "" {
			writePage(w, "Checkout", buildCheckoutPage(items, a.products, "Please provide a name and payment details.", ""))
			return
		}
		orderID := fmt.Sprintf("NS-%06d", rand.Intn(1000000))
		a.clearCart(cartID)
		writePage(w, "Order confirmed", buildOrderCompletePage(orderID, a.currentUser(r), cartTotal(items, a.products)))
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (a *app) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{Name: "auth_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *app) requestOTP(identifier string) (string, error) {
	if strings.TrimSpace(identifier) == "" {
		return "", fmt.Errorf("missing identifier")
	}
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	a.otpStore[identifier] = otpEntry{Code: code, ExpiresAt: time.Now().Add(5 * time.Minute)}
	return code, nil
}

func (a *app) rateLimit(key string) bool {
	return a.rl.allow(key)
}

func (a *app) verifyOTP(identifier, code string) (string, error) {
	entry, ok := a.otpStore[identifier]
	if !ok {
		return "", fmt.Errorf("no active OTP found for %s", identifier)
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(a.otpStore, identifier)
		return "", fmt.Errorf("OTP expired")
	}
	if entry.Code != code {
		return "", fmt.Errorf("invalid OTP")
	}
	delete(a.otpStore, identifier)
	return a.issueJWT(identifier)
}

func (a *app) issueJWT(subject string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"%s","exp":%d}`, subject, time.Now().Add(24*time.Hour).Unix())))
	signingInput := header + "." + payload
	signature := hmacSHA256(signingInput, []byte(jwtSecret))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (a *app) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/login") || r.URL.Path == "/" || r.URL.Path == "/cart" || r.URL.Path == "/cart/add" || r.URL.Path == "/cart/remove" {
			next.ServeHTTP(w, r)
			return
		}
		token, err := r.Cookie("auth_token")
		if err != nil || !a.validJWT(token.Value) {
			http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *app) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.rl.allow(getClientIP(r)) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *app) validJWT(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	if !hmac.Equal([]byte(parts[2]), []byte(base64.RawURLEncoding.EncodeToString(hmacSHA256(parts[0]+"."+parts[1], []byte(jwtSecret))))) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}
	return time.Unix(int64(exp), 0).After(time.Now())
}

func (a *app) currentUser(r *http.Request) string {
	token, err := r.Cookie("auth_token")
	if err != nil || !a.validJWT(token.Value) {
		return "Guest"
	}
	parts := strings.Split(token.Value, ".")
	if len(parts) != 3 {
		return "Guest"
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "Guest"
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "Guest"
	}
	if sub, ok := claims["sub"].(string); ok {
		return sub
	}
	return "Guest"
}

func (a *app) getCartID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("cart_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	cartID := fmt.Sprintf("cart-%d", rand.Intn(1000000))
	http.SetCookie(w, &http.Cookie{Name: "cart_id", Value: cartID, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return cartID
}

func (a *app) ensureCart(cartID string) map[string]int {
	cart, ok := a.carts[cartID]
	if !ok {
		cart = make(map[string]int)
		a.carts[cartID] = cart
	}
	return cart
}

func (a *app) cartItems(cartID string) map[string]int {
	cart, ok := a.carts[cartID]
	if !ok {
		return map[string]int{}
	}
	return cart
}

func (a *app) cartCount(cartID string) int {
	cart := a.cartItems(cartID)
	count := 0
	for _, qty := range cart {
		count += qty
	}
	return count
}

func (a *app) clearCart(cartID string) {
	delete(a.carts, cartID)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func hmacSHA256(message string, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(message))
	return h.Sum(nil)
}

func writePage(w http.ResponseWriter, title, body string) {
	_, _ = io.WriteString(w, fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <style>
    :root { color-scheme: dark; }
    body { font-family: Arial, sans-serif; margin: 0; background: #0f172a; color: #f8fafc; }
    a { color: #7dd3fc; }
    header { background: linear-gradient(135deg, #111827, #1d4ed8); padding: 24px 32px; }
    nav { display:flex; justify-content:space-between; align-items:center; gap: 16px; }
    main { max-width: 1120px; margin: 0 auto; padding: 32px; }
    .grid { display:grid; gap: 20px; grid-template-columns: repeat(auto-fit, minmax(240px,1fr)); }
    .card { background:#111827; border:1px solid #334155; border-radius:16px; padding:20px; box-shadow: 0 10px 30px rgba(0,0,0,0.3); }
    .pill { display:inline-block; background:#1d4ed8; color:white; padding:6px 12px; border-radius:999px; margin-bottom:14px; font-size: 13px; }
    form { display:flex; flex-direction:column; gap:10px; }
    input, button, textarea { padding: 12px 14px; border-radius: 10px; border: 1px solid #475569; background: #020617; color: white; }
    button { cursor:pointer; background:#38bdf8; color:#082f49; font-weight:700; }
    .muted { color:#94a3b8; }
    .summary { background:#0f172a; border:1px solid #334155; border-radius:16px; padding:20px; }
    .row { display:flex; justify-content:space-between; align-items:center; gap:10px; margin:8px 0; }
    .actions { display:flex; gap:10px; flex-wrap:wrap; }
  </style>
</head>
<body>
  <header>
    <nav>
      <div><strong>Northstar Studio</strong><br><span class="muted">Modern essentials for bold wardrobes</span></div>
      <div class="actions">
        <a href="/">Shop</a>
        <a href="/cart">Cart</a>
        <a href="/checkout">Checkout</a>
      </div>
    </nav>
  </header>
  <main>%s</main>
</body>
</html>`, title, body))
}

func buildHomePage(cartCount int, user string, products []product) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<section class="card"><span class="pill">New season drop</span><h1>Build your signature wardrobe.</h1><p class="muted">High-performance pieces with premium comfort and secure checkout.</p><div class="actions"><a href="/checkout">Proceed to checkout</a> <a href="/login">%s</a></div></section>`, user)
	fmt.Fprintf(&b, `<section class="grid">`)
	for _, p := range products {
		fmt.Fprintf(&b, `<div class="card"><h3>%s</h3><p class="muted">%s</p><div class="row"><strong>$%d</strong><form action="/cart/add" method="post"><input type="hidden" name="product_id" value="%s"><button type="submit">Add to cart</button></form></div></div>`, p.Name, p.Description, p.PriceCents/100, p.ID)
	}
	fmt.Fprintf(&b, `</section><p class="muted">Cart count: %d</p>`, cartCount)
	return b.String()
}

func buildCartPage(items map[string]int, products []product, user string, cartID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<section class="card"><h2>Your cart</h2><p class="muted">Signed in as %s</p>`, user)
	if len(items) == 0 {
		b.WriteString(`<p>Your cart is empty. Start with one of the signature pieces above.</p>`)
		return b.String() + `</section>`
	}
	b.WriteString(`<div class="summary">`)
	total := 0
	for _, p := range products {
		if qty := items[p.ID]; qty > 0 {
			line := qty * p.PriceCents
			total += line
			fmt.Fprintf(&b, `<div class="row"><span>%s × %d</span><span>$%d</span><form action="/cart/remove" method="post" style="display:inline"><input type="hidden" name="product_id" value="%s"><button type="submit">Remove</button></form></div>`, p.Name, qty, line/100, p.ID)
		}
	}
	fmt.Fprintf(&b, `<div class="row"><strong>Total</strong><strong>$%d</strong></div>`, total/100)
	b.WriteString(`<div class="actions"><a href="/checkout">Checkout securely</a> <a href="/">Continue shopping</a></div></div></section>`)
	return b.String()
}

func buildLoginPage(identifier, message, code string) string {
	var b strings.Builder
	b.WriteString(`<section class="grid"><div class="card"><h2>Sign in with OTP</h2><p class="muted">Enter your email or phone number to receive a one-time code.</p><form action="/login/request-otp" method="post"><input name="identifier" value="` + identifier + `" placeholder="you@example.com or +1-555-5555"><button type="submit">Request OTP</button></form></div><div class="card"><h3>Verify code</h3><form action="/login/verify-otp" method="post"><input name="identifier" value="` + identifier + `" placeholder="same email or phone"><input name="code" placeholder="6-digit code"><button type="submit">Verify and continue</button></form></div></section>`)
	if message != "" {
		fmt.Fprintf(&b, `<p class="pill">%s</p>`, message)
	}
	if code != "" {
		fmt.Fprintf(&b, `<p class="muted">Demo OTP: %s</p>`, code)
	}
	return b.String()
}

func buildCheckoutPage(items map[string]int, products []product, message, hidden string) string {
	var b strings.Builder
	b.WriteString(`<section class="grid"><div class="card"><h2>Secure checkout</h2><p class="muted">Your address and payment details stay protected behind a signed session.</p><form action="/checkout" method="post"><input name="name" placeholder="Full name"><input name="card" placeholder="Card ending in 4242"><textarea name="address" placeholder="Shipping address"></textarea><button type="submit">Pay now</button></form></div><div class="summary"><h3>Order summary</h3>`)
	if message != "" {
		fmt.Fprintf(&b, `<p class="pill">%s</p>`, message)
	}
	total := 0
	for _, p := range products {
		if qty := items[p.ID]; qty > 0 {
			line := qty * p.PriceCents
			total += line
			fmt.Fprintf(&b, `<div class="row"><span>%s × %d</span><span>$%d</span></div>`, p.Name, qty, line/100)
		}
	}
	fmt.Fprintf(&b, `<div class="row"><strong>Total</strong><strong>$%d</strong></div></div></section>`, total/100)
	if hidden != "" {
		fmt.Fprintf(&b, `<p class="muted">%s</p>`, hidden)
	}
	return b.String()
}

func buildOrderCompletePage(orderID, user string, total int) string {
	return fmt.Sprintf(`<section class="card"><span class="pill">Payment secured</span><h2>Order confirmed</h2><p>Hi %s, your order %s has been placed successfully.</p><p class="muted">Total charged: $%d</p><p><a href="/">Return to the catalog</a></p></section>`, user, orderID, total)
}

func cartTotal(items map[string]int, products []product) int {
	total := 0
	for _, p := range products {
		total += items[p.ID] * p.PriceCents
	}
	return total / 100
}
