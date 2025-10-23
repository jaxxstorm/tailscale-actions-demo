-- Remove index
DROP INDEX IF EXISTS idx_products_category;

-- Remove products added in 002
DELETE FROM products WHERE name = 'Tailscale Enterprise';

-- Remove columns added in 002
ALTER TABLE products DROP COLUMN IF EXISTS category;
ALTER TABLE products DROP COLUMN IF EXISTS stock_quantity;
