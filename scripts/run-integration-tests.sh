#!/usr/bin/env bash
set -euo pipefail

# Script to run integration tests with credentials from AWS Secrets Manager
# This script:
# 1. Gets database credentials from AWS Secrets Manager
# 2. Sets environment variables for tests
# 3. Runs Go integration tests

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

# Set environment variables for tests
export DB_HOST="$DB_HOST"
export DB_PORT="5432"
export DB_USER="$DB_USER"
export DB_PASSWORD="$DB_PASSWORD"
export DB_NAME="$DB_NAME"
export DB_SSLMODE="require"

# Check if TEST_API_URL is already set, otherwise use default
if [ -z "${TEST_API_URL:-}" ]; then
    echo ""
    echo "⚠️  TEST_API_URL not set. Using default: http://demo:8080"
    echo "   Set TEST_API_URL environment variable to override."
    export TEST_API_URL="http://demo:8080"
else
    echo ""
    echo "✅ Using TEST_API_URL: $TEST_API_URL"
fi

echo ""
echo "🧪 Running integration tests..."
echo ""

# Run tests from app directory
cd app
go test -v -count=1 ./...

echo ""
echo "✅ Integration tests complete!"
