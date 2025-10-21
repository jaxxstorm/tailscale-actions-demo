# Database Migrations

This directory contains versioned database migrations organized by deployment stage.

## Structure

```
migrations/
├── initial/          # Initial database setup (v001)
├── new_product/      # Feature addition (v002)
└── rollback/         # Feature rollback (v003)
```

## Stages

### 1. Initial (`migrations/initial/`)

The baseline database schema with minimal data.

**What it does:**
- Creates the `products` table with basic columns (id, name, description, price, timestamps)
- Sets up indexes and triggers for `updated_at`
- Inserts 2 initial products:
  - Tailscale Personal ($0.00)
  - Tailscale Premium ($6.00)

**Run with:**
```bash
make migrate-initial
# or
migrate -path ./migrations/initial -database "postgres://postgres:postgres@localhost:5432/demo?sslmode=disable" up
```

### 2. New Product (`migrations/new_product/`)

Adds a new product tier and extends the schema.

**What it does:**
- Adds `stock_quantity` and `category` columns
- Updates existing products with stock and category data
- Inserts Tailscale Enterprise product ($15.00)
- Creates index on category column

**Run with:**
```bash
make migrate-new-product
# or
migrate -path ./migrations/new_product -database "postgres://postgres:postgres@localhost:5432/demo?sslmode=disable" up
```

### 3. Rollback (`migrations/rollback/`)

Demonstrates a feature rollback scenario.

**What it does:**
- Removes the Tailscale Enterprise product
- Keeps the schema changes (columns remain)
- Simulates rolling back a problematic feature

**Run with:**
```bash
make migrate-rollback
# or
migrate -path ./migrations/rollback -database "postgres://postgres:postgres@localhost:5432/demo?sslmode=disable" up
```

## Usage

### Sequential Deployment

Run migrations in order to simulate a full deployment cycle:

```bash
# Start fresh
make migrate-initial

# Deploy new feature
make migrate-new-product

# Rollback if needed
make migrate-rollback
```

## Demo Scenarios

### Scenario 1: Clean Deployment
```bash
make migrate-initial
make migrate-new-product
# Result: 3 products with full schema
```

### Scenario 2: Rollback After Issues
```bash
make migrate-initial
make migrate-new-product
make migrate-rollback
# Result: 2 products with full schema (Enterprise removed)
```

### Scenario 3: Fresh Start
```bash
make migrate-down STAGE=initial
make migrate-initial
# Result: Clean slate with 2 products
```

## Version Numbers

Each migration file has a version prefix:
- `001_*` - Initial setup
- `002_*` - Feature addition  
- `003_*` - Rollback

This ensures proper ordering when running migrations.
