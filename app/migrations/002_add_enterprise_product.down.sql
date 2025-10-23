-- Remove CI/CD product added in migration 002
DELETE FROM products WHERE name = 'CI/CD';
