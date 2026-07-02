# AARYA — Storefront (Next.js)

Responsive storefront + admin console for the AARYA clothing brand, talking to the Go API in this repo.

## Stack
- Next.js 14 (App Router) + TypeScript
- Tailwind CSS (design tokens in `tailwind.config.ts` / `app/globals.css`)
- Fonts: Fraunces (display) + Jost (body)

## Setup
```bash
cd frontend
npm install
cp .env.local.example .env.local   # set NEXT_PUBLIC_API_URL if the API isn't on :8080
npm run dev                        # http://localhost:3000
```

The Go API must be running (default `http://localhost:8080`) and its
`ALLOWED_ORIGINS` must include `http://localhost:3000`.

## Routes
| Path | Purpose |
|------|---------|
| `/` | Home — hero, categories, new arrivals, bestsellers |
| `/shop` | Listing with category filter + sort (`?category=slug`) |
| `/product/[slug]` | Product detail, variant/size, add to bag |
| `/cart` | Bag + coupon + summary |
| `/checkout` | OTP verify + shipping + payment (COD / Razorpay) |
| `/account` | OTP login + order history |
| `/admin` | Inventory (add product, edit stock, deactivate) — vendor/admin only |
| `/admin/orders` | Order fulfilment + status updates |

## Notes
- Auth is OTP-based; in dev the API returns the OTP in the response and the UI shows it.
- Cart is guest/session based via an `X-Session-ID` header (persisted in `localStorage`).
- Razorpay is wired as an order option (payment stays `pending`); the hosted
  checkout popup (`checkout.js`) can be added on top when keys are configured.
