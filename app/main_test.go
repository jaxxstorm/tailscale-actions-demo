package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestConfig holds test configuration
type TestConfig struct {
	APIBaseURL string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// getTestConfig reads test configuration from environment
func getTestConfig() TestConfig {
	return TestConfig{
		APIBaseURL: getEnvOrDefault("TEST_API_URL", "http://localhost:8080"),
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "postgres"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:     getEnvOrDefault("DB_NAME", "demo"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// waitForServer waits for the API server to be ready
func waitForServer(baseURL string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server to be ready")
		case <-ticker.C:
			resp, err := http.Get(baseURL + "/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

// TestHealth tests the health endpoint
func TestHealth(t *testing.T) {
	config := getTestConfig()

	// Wait for server to be ready
	if err := waitForServer(config.APIBaseURL, 30*time.Second); err != nil {
		t.Fatalf("Server not ready: %v", err)
	}

	resp, err := http.Get(config.APIBaseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if health.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", health.Status)
	}

	if health.Database != "connected" {
		t.Errorf("Expected database 'connected', got '%s'", health.Database)
	}

	t.Logf("Health check passed: %+v", health)
}

// TestUserEndpoint tests the user API endpoint
func TestUserEndpoint(t *testing.T) {
	config := getTestConfig()

	resp, err := http.Get(config.APIBaseURL + "/api/user")
	if err != nil {
		t.Fatalf("Failed to call user endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		t.Fatalf("Failed to decode user response: %v", err)
	}

	// User may or may not be connected via Tailscale
	t.Logf("User info: connected=%v, login=%s, display=%s",
		userInfo.Connected, userInfo.LoginName, userInfo.DisplayName)
}

// TestProductsEndpoint tests the products API endpoint
func TestProductsEndpoint(t *testing.T) {
	config := getTestConfig()

	resp, err := http.Get(config.APIBaseURL + "/api/products")
	if err != nil {
		t.Fatalf("Failed to call products endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var products []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		t.Fatalf("Failed to decode products response: %v", err)
	}

	t.Logf("Retrieved %d products", len(products))

	// Verify products have expected structure
	for i, p := range products {
		// Check for required fields
		if _, ok := p["id"]; !ok {
			t.Errorf("Product %d missing 'id' field", i)
		}
		if name, ok := p["name"]; !ok || name == "" {
			t.Errorf("Product %d missing or empty 'name' field", i)
		}
		if _, ok := p["price"]; !ok {
			t.Errorf("Product %d missing 'price' field", i)
		}
		if _, ok := p["created_at"]; !ok {
			t.Errorf("Product %d missing 'created_at' field", i)
		}
	}
}

// TestProductsWithSeededData tests that seeded data is accessible
func TestProductsWithSeededData(t *testing.T) {
	config := getTestConfig()

	// First, verify we can connect to the database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count products in database
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count products: %v", err)
	}

	t.Logf("Database has %d products", count)

	// Now test the API
	resp, err := http.Get(config.APIBaseURL + "/api/products")
	if err != nil {
		t.Fatalf("Failed to call products endpoint: %v", err)
	}
	defer resp.Body.Close()

	var products []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		t.Fatalf("Failed to decode products response: %v", err)
	}

	if len(products) != count {
		t.Errorf("Expected %d products from API, got %d", count, len(products))
	}

	// If we have products, verify the first one
	if len(products) > 0 {
		name := products[0]["name"]
		price := products[0]["price"]
		t.Logf("First product: %v - $%v", name, price)
	}
}

// TestDatabaseConnection tests direct database connectivity
func TestDatabaseConnection(t *testing.T) {
	config := getTestConfig()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Verify products table exists
	var tableName string
	err = db.QueryRowContext(ctx,
		"SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename='products'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Products table does not exist: %v", err)
	}

	t.Logf("Successfully verified database and products table")
}
