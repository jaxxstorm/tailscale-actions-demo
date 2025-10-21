-- Re-add Enterprise product
INSERT INTO products (name, description, price, stock_quantity, category) VALUES
    ('Tailscale Enterprise', 'Full-featured secure networking for organizations', 15.00, 25, 'Subscription')
ON CONFLICT DO NOTHING;
