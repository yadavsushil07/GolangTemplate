-- Rollback 005_attributes

DROP TABLE IF EXISTS product_attribute_values;
DROP TABLE IF EXISTS attribute_values;
DROP TABLE IF EXISTS attributes;

-- Restore original CASCADE behaviour on category FK
ALTER TABLE product_categories DROP CONSTRAINT IF EXISTS product_categories_category_id_fkey;
ALTER TABLE product_categories ADD CONSTRAINT product_categories_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE;
