#!/usr/bin/env bash
set -euo pipefail

# Script to query products from the database
# This script:
# 1. Gets database credentials from AWS Secrets Manager
# 2. Queries all products from the database

cd "$(dirname "$0")/.."

echo "🔍 Getting database credentials from AWS Secrets Manager..."

# Get the secret ARN from Terraform output
DB_SECRET_ARN=$(cd infra && terraform output -raw db_secret_arn)
echo "Secret ARN: $DB_SECRET_ARN"

# Get the entire secret as JSON from Secrets Manager
SECRET_JSON=$(aws secretsmanager get-secret-value \
    --secret-id "$DB_SECRET_ARN" \
    --query SecretString \
    --output text)

# Parse the JSON to get individual values
DB_HOST=$(echo "$SECRET_JSON" | jq -r '.host')
DB_NAME=$(echo "$SECRET_JSON" | jq -r '.dbname')
DB_USER=$(echo "$SECRET_JSON" | jq -r '.username')
DB_PASSWORD=$(echo "$SECRET_JSON" | jq -r '.password')

echo "✅ Retrieved database credentials"
echo "DB Host: $DB_HOST"
echo "DB Name: $DB_NAME"
echo "DB User: $DB_USER"

# Set password for psql
export PGPASSWORD="$DB_PASSWORD"

echo ""
echo "📦 Products in database:"
echo ""
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT id, name, price, stock_quantity, category, created_at FROM products ORDER BY name;"

echo ""
echo "📊 Product count:"
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) as total_products FROM products;"

echo ""
echo "✅ Done!"
