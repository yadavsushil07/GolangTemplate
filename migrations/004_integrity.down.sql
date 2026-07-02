-- Revert 004

-- order_items
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS chk_order_items_price_nonneg;
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS chk_order_items_qty_pos;
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_product_id_fkey;
ALTER TABLE order_items ADD CONSTRAINT order_items_product_id_fkey
	FOREIGN KEY (product_id) REFERENCES products(id);
ALTER TABLE order_items ALTER COLUMN product_id DROP NOT NULL;
ALTER TABLE order_items ALTER COLUMN order_id DROP NOT NULL;

-- orders
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_total_nonneg;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_user_id_fkey;
ALTER TABLE orders ADD CONSTRAINT orders_user_id_fkey
	FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE orders ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE orders DROP COLUMN IF EXISTS updated_at;

-- cart_items
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS chk_cart_qty_pos;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_variant_id_fkey;
ALTER TABLE cart_items ADD CONSTRAINT cart_items_variant_id_fkey
	FOREIGN KEY (variant_id) REFERENCES product_variants(id);
ALTER TABLE cart_items ALTER COLUMN product_id DROP NOT NULL;

-- product_variants
DROP INDEX IF EXISTS idx_variants_sku;
ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS chk_variants_stock_nonneg;
ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS chk_variants_price_nonneg;

-- coupons
ALTER TABLE coupons DROP CONSTRAINT IF EXISTS chk_coupons_one_discount;

-- products
ALTER TABLE products DROP CONSTRAINT IF EXISTS chk_products_stock_nonneg;
ALTER TABLE products DROP CONSTRAINT IF EXISTS chk_products_price_nonneg;
DROP INDEX IF EXISTS idx_products_slug;
ALTER TABLE products DROP COLUMN IF EXISTS slug;

-- supporting indexes
DROP INDEX IF EXISTS idx_cart_items_variant;
DROP INDEX IF EXISTS idx_order_items_product;
DROP INDEX IF EXISTS idx_order_items_variant;
DROP INDEX IF EXISTS idx_product_categories_category;
