-- Add new CI/CD product
INSERT INTO products (name, description, price, stock_quantity, category) 
VALUES ('CI/CD', 'Secure CI/CD pipelines and build infrastructure', 179.00, 40, 'DevOps')
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock_quantity = EXCLUDED.stock_quantity,
    category = EXCLUDED.category;
