-- Create products table with unique constraint on name
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INTEGER DEFAULT 0,
    category VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);

-- Create index on category for filtering
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- DECLARATIVE: Ensure exactly these products exist (5 initial SKUs)
-- Delete any products that shouldn't be here
DELETE FROM products WHERE name NOT IN ('Business VPN', 'Workload Connectivity', 'Edge & IoT', 'Securing AI', 'Homelab');

-- Upsert Business VPN
INSERT INTO products (name, description, price, stock_quantity, category) 
VALUES ('Business VPN', 'Secure remote access for distributed teams', 99.00, 50, 'Networking')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock_quantity = EXCLUDED.stock_quantity,
    category = EXCLUDED.category;

-- Upsert Workload Connectivity
INSERT INTO products (name, description, price, stock_quantity, category) 
VALUES ('Workload Connectivity', 'Secure service-to-service communication', 149.00, 75, 'Infrastructure')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock_quantity = EXCLUDED.stock_quantity,
    category = EXCLUDED.category;

-- Upsert Edge & IoT
INSERT INTO products (name, description, price, stock_quantity, category) 
VALUES ('Edge & IoT', 'Connect edge devices and IoT infrastructure', 199.00, 30, 'IoT')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock_quantity = EXCLUDED.stock_quantity,
    category = EXCLUDED.category;

-- Upsert Securing AI
INSERT INTO products (name, description, price, stock_quantity, category) 
VALUES ('Securing AI', 'Secure AI workloads and model training', 299.00, 20, 'AI/ML')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock_quantity = EXCLUDED.stock_quantity,
    category = EXCLUDED.category;

-- Upsert Homelab
INSERT INTO products (name, description, price, stock_quantity, category) 
VALUES ('Homelab', 'Personal networking for hobbyists and tinkerers', 0.00, 200, 'Personal')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock_quantity = EXCLUDED.stock_quantity,
    category = EXCLUDED.category;
