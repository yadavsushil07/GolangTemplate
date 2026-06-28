ALTER TABLE order_items DROP COLUMN IF EXISTS variant_id;
ALTER TABLE cart_items  DROP COLUMN IF EXISTS variant_id;
ALTER TABLE orders
    DROP COLUMN IF EXISTS payment_method,
    DROP COLUMN IF EXISTS payment_status,
    DROP COLUMN IF EXISTS razorpay_order_id,
    DROP COLUMN IF EXISTS coupon_code,
    DROP COLUMN IF EXISTS discount_cents,
    DROP COLUMN IF EXISTS customization_note;

DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS categories;
