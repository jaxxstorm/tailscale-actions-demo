#!/usr/bin/env bash
set -euo pipefail

# Script to reset the database using migrations
# This script:
# 1. Gets database credentials from AWS Secrets Manager
# 2. Runs migrate commands to reset the database

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

# URL encode the password
ENCODED_PASSWORD=$(printf %s "$DB_PASSWORD" | jq -sRr @uri)

# Construct database URL
DB_URL="postgres://${DB_USER}:${ENCODED_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=require"

echo ""
echo "🗑️  Dropping all tables and schema..."
migrate -path app/migrations -database "$DB_URL" drop -f

echo ""
echo "🔄 Running migrations to version 1..."
migrate -path app/migrations -database "$DB_URL" goto 1

echo ""
echo "✅ Database reset complete!"
echo ""
echo "📊 Checking final state..."

# Use psql to verify
export PGPASSWORD="$DB_PASSWORD"
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version, dirty FROM schema_migrations;"
echo ""
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT name, price, category FROM products ORDER BY name;"

echo ""
echo "✅ Done! Database is at version 1 with 5 products."
