-- 004: data-integrity hardening — product slugs, NOT NULL, ON DELETE rules, CHECK constraints, indexes.
-- Retrofitted CHECK constraints use NOT VALID so the migration never fails on
-- pre-existing rows; all new/updated rows are validated going forward.

-- ============ products ============
ALTER TABLE products ADD COLUMN IF NOT EXISTS slug TEXT;

UPDATE products
SET slug = trim(both '-' from regexp_replace(lower(coalesce(name, 'product')), '[^a-z0-9]+', '-', 'g')) || '-' || id
WHERE slug IS NULL OR slug = '';

ALTER TABLE products ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_slug ON products(slug);

DO $$ BEGIN
	ALTER TABLE products ADD CONSTRAINT chk_products_price_nonneg CHECK (price_cents >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
	ALTER TABLE products ADD CONSTRAINT chk_products_stock_nonneg CHECK (stock >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ============ product_variants ============
DO $$ BEGIN
	ALTER TABLE product_variants ADD CONSTRAINT chk_variants_price_nonneg CHECK (price_cents >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
	ALTER TABLE product_variants ADD CONSTRAINT chk_variants_stock_nonneg CHECK (stock >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
CREATE UNIQUE INDEX IF NOT EXISTS idx_variants_sku ON product_variants(sku) WHERE sku IS NOT NULL AND sku <> '';

-- ============ cart_items ============
ALTER TABLE cart_items ALTER COLUMN product_id SET NOT NULL;
DO $$ BEGIN
	ALTER TABLE cart_items ADD CONSTRAINT chk_cart_qty_pos CHECK (quantity > 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_variant_id_fkey;
ALTER TABLE cart_items ADD CONSTRAINT cart_items_variant_id_fkey
	FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE CASCADE;

-- ============ orders ============
ALTER TABLE orders ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE orders ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_user_id_fkey;
ALTER TABLE orders ADD CONSTRAINT orders_user_id_fkey
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
DO $$ BEGIN
	ALTER TABLE orders ADD CONSTRAINT chk_orders_total_nonneg CHECK (total_cents >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ============ order_items ============
ALTER TABLE order_items ALTER COLUMN order_id SET NOT NULL;
ALTER TABLE order_items ALTER COLUMN product_id SET NOT NULL;
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_product_id_fkey;
ALTER TABLE order_items ADD CONSTRAINT order_items_product_id_fkey
	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT;
DO $$ BEGIN
	ALTER TABLE order_items ADD CONSTRAINT chk_order_items_qty_pos CHECK (quantity > 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
	ALTER TABLE order_items ADD CONSTRAINT chk_order_items_price_nonneg CHECK (price_cents >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ============ coupons ============
-- Exactly one of discount_pct / discount_cents must be set.
DO $$ BEGIN
	ALTER TABLE coupons ADD CONSTRAINT chk_coupons_one_discount
		CHECK ((discount_pct IS NOT NULL) <> (discount_cents IS NOT NULL)) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ============ supporting indexes ============
CREATE INDEX IF NOT EXISTS idx_cart_items_variant ON cart_items(variant_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_order_items_variant ON order_items(variant_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_category ON product_categories(category_id);
