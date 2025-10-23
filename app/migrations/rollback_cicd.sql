-- Manual Rollback Script
-- Run this script manually to remove the CI/CD product

-- Remove the CI/CD product
DELETE FROM products WHERE name = 'CI/CD';

-- Verify the change
SELECT name, price, category FROM products ORDER BY name;
