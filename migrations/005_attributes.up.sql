-- 005: product attribute system + category delete protection

-- 1. Fix category delete: change CASCADE to RESTRICT so a category with
--    products assigned cannot be deleted.
ALTER TABLE product_categories DROP CONSTRAINT IF EXISTS product_categories_category_id_fkey;
ALTER TABLE product_categories ADD CONSTRAINT product_categories_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;

-- 2. Attribute definitions (Size, Colour, Material, Occasion, …)
CREATE TABLE IF NOT EXISTS attributes (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 3. Allowed values per attribute (XS, S, M … / Red, Blue … / Cotton …)
CREATE TABLE IF NOT EXISTS attribute_values (
    id           BIGSERIAL PRIMARY KEY,
    attribute_id BIGINT NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    value        TEXT NOT NULL,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    UNIQUE(attribute_id, value)
);

-- 4. Link products to chosen attribute values (many-to-many)
CREATE TABLE IF NOT EXISTS product_attribute_values (
    product_id         BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    attribute_value_id BIGINT NOT NULL REFERENCES attribute_values(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, attribute_value_id)
);

CREATE INDEX IF NOT EXISTS idx_pav_product   ON product_attribute_values(product_id);
CREATE INDEX IF NOT EXISTS idx_pav_av        ON product_attribute_values(attribute_value_id);
CREATE INDEX IF NOT EXISTS idx_attr_values_attr ON attribute_values(attribute_id);

-- 5. Seed standard attribute groups
INSERT INTO attributes (name, sort_order) VALUES
    ('Size',     1),
    ('Colour',   2),
    ('Material', 3),
    ('Occasion', 4)
ON CONFLICT (name) DO NOTHING;

-- 6. Seed Size values
INSERT INTO attribute_values (attribute_id, value, sort_order)
SELECT a.id, v.val, v.ord
FROM attributes a
JOIN (VALUES
    ('XS',1),('S',2),('M',3),('L',4),('XL',5),
    ('2XL',6),('3XL',7),('4XL',8),('5XL',9),('6XL',10),
    ('Free Size',11)
) AS v(val,ord) ON TRUE
WHERE a.name = 'Size'
ON CONFLICT (attribute_id, value) DO NOTHING;

-- 7. Seed Colour values
INSERT INTO attribute_values (attribute_id, value, sort_order)
SELECT a.id, v.val, v.ord
FROM attributes a
JOIN (VALUES
    ('Red',1),('Pink',2),('Blue',3),('Green',4),('Yellow',5),
    ('Orange',6),('Purple',7),('Black',8),('White',9),('Beige',10),
    ('Maroon',11),('Navy',12),('Gold',13),('Silver',14),('Multi',15)
) AS v(val,ord) ON TRUE
WHERE a.name = 'Colour'
ON CONFLICT (attribute_id, value) DO NOTHING;

-- 8. Seed Material values
INSERT INTO attribute_values (attribute_id, value, sort_order)
SELECT a.id, v.val, v.ord
FROM attributes a
JOIN (VALUES
    ('Cotton',1),('Silk',2),('Georgette',3),('Chiffon',4),('Crepe',5),
    ('Net',6),('Velvet',7),('Linen',8),('Rayon',9),('Polyester',10)
) AS v(val,ord) ON TRUE
WHERE a.name = 'Material'
ON CONFLICT (attribute_id, value) DO NOTHING;

-- 9. Seed Occasion values
INSERT INTO attribute_values (attribute_id, value, sort_order)
SELECT a.id, v.val, v.ord
FROM attributes a
JOIN (VALUES
    ('Casual',1),('Party',2),('Wedding',3),('Festival',4),
    ('Office',5),('Ethnic',6),('Bridal',7)
) AS v(val,ord) ON TRUE
WHERE a.name = 'Occasion'
ON CONFLICT (attribute_id, value) DO NOTHING;
