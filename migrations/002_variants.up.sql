-- Categories / Collections
CREATE TABLE IF NOT EXISTS categories (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    slug       TEXT UNIQUE NOT NULL,
    sort_order INTEGER DEFAULT 0
);

-- Product → Category (many-to-many)
CREATE TABLE IF NOT EXISTS product_categories (
    product_id  BIGINT REFERENCES products(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

-- Product Variants (size + color/fabric, each with its own price)
CREATE TABLE IF NOT EXISTS product_variants (
    id          BIGSERIAL PRIMARY KEY,
    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    size        TEXT NOT NULL,
    color       TEXT,
    price_cents INTEGER NOT NULL,
    stock       INTEGER NOT NULL DEFAULT 0,
    sku         TEXT,
    is_active   BOOLEAN DEFAULT TRUE
);

-- Multiple product images
CREATE TABLE IF NOT EXISTS product_images (
    id         BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0
);

-- Coupon / Discount codes
CREATE TABLE IF NOT EXISTS coupons (
    id              BIGSERIAL PRIMARY KEY,
    code            TEXT UNIQUE NOT NULL,
    discount_pct    INTEGER,          -- e.g. 15 = 15% off (exclusive with discount_cents)
    discount_cents  INTEGER,          -- flat amount off
    min_order_cents INTEGER DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Extend orders with payment + coupon + customization
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS payment_method    TEXT NOT NULL DEFAULT 'cod',
    ADD COLUMN IF NOT EXISTS payment_status    TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS razorpay_order_id TEXT,
    ADD COLUMN IF NOT EXISTS coupon_code       TEXT,
    ADD COLUMN IF NOT EXISTS discount_cents    INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS customization_note TEXT;

-- Link order_items to a variant
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS variant_id BIGINT REFERENCES product_variants(id);

-- Replace cart_items session+product unique constraint with session+variant
ALTER TABLE cart_items
    ADD COLUMN IF NOT EXISTS variant_id BIGINT REFERENCES product_variants(id);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_product_variants_product ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_images_product   ON product_images(product_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_product ON product_categories(product_id);
