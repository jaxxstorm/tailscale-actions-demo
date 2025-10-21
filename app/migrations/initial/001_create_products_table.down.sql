-- Drop products table and related objects
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_products_created_at;
DROP TABLE IF EXISTS products;
