package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	_ "github.com/lib/pq"
	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

type Server struct {
	db        *sql.DB
	client    *tailscale.LocalClient
	tsnetMode bool
}

type UserInfo struct {
	Connected    bool   `json:"connected"`
	LoginName    string `json:"login_name,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	FirstInitial string `json:"first_initial,omitempty"`
	Error        string `json:"error,omitempty"`
}

type WhoIsData struct {
	LoginName   string
	DisplayName string
}

type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Tailscale string `json:"tailscale"`
}

type Config struct {
	DBHost            string `env:"DB_HOST" default:"localhost" help:"Database host"`
	DBPort            string `env:"DB_PORT" default:"5432" help:"Database port"`
	DBUser            string `env:"DB_USER" default:"postgres" help:"Database user"`
	DBPassword        string `env:"DB_PASSWORD" default:"postgres" help:"Database password"`
	DBName            string `env:"DB_NAME" default:"demo" help:"Database name"`
	Port              string `env:"PORT" default:"8080" help:"HTTP server port"`
	TailscaleAuthKey  string `env:"TS_AUTHKEY" help:"Tailscale auth key for tsnet mode"`
	TailscaleHostname string `env:"TS_HOSTNAME" default:"demo" help:"Hostname for tsnet registration"`
}

func main() {
	// Parse configuration using kong
	var config Config
	kong.Parse(&config,
		kong.Name("tailscale-demo"),
		kong.Description("Tailscale demo application with PostgreSQL integration"),
		kong.UsageOnError(),
	)

	// Initialize database connection
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Printf("Warning: Failed to ping database: %v", err)
	} else {
		log.Println("Successfully connected to database")
	}

	// Determine if we're running in tsnet mode
	useTsnet := config.TailscaleAuthKey != ""

	// Create server instance (client will be set in tsnet mode)
	server := &Server{
		db:        db,
		client:    &tailscale.LocalClient{},
		tsnetMode: useTsnet,
	}

	// Setup HTTP handlers
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve index.html at root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}
		http.NotFound(w, r)
	})

	// API endpoints
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/api/user", server.userHandler)
	mux.HandleFunc("/api/products", server.productsHandler)

	// Start server based on mode
	if useTsnet {
		log.Printf("Starting in tsnet mode with hostname: %s", config.TailscaleHostname)
		startTsnetServer(config, server, mux)
	} else {
		log.Printf("Starting in regular HTTP mode on port %s", config.Port)
		startRegularServer(config, mux)
	}
}

func startTsnetServer(config Config, server *Server, handler http.Handler) {
	ts := &tsnet.Server{
		Hostname: config.TailscaleHostname,
		AuthKey:  config.TailscaleAuthKey,
		Logf:     log.Printf,
	}

	defer ts.Close()

	// Start the tsnet server
	if err := ts.Start(); err != nil {
		log.Fatalf("Failed to start tsnet server: %v", err)
	}

	// Update the server to use tsnet's LocalClient
	lc, err := ts.LocalClient()
	if err != nil {
		log.Fatalf("Failed to get tsnet LocalClient: %v", err)
	}
	server.client = lc

	log.Printf("Tailscale node started successfully")

	// Listen on the configured port (default 80 for HTTP, but use config.Port)
	listenAddr := fmt.Sprintf(":%s", config.Port)
	ln, err := ts.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
	}
	defer ln.Close()

	httpServer := &http.Server{
		Handler: handler,
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on Tailscale network")
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down tsnet server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func startRegularServer(config Config, handler http.Handler) {
	httpServer := &http.Server{
		Addr:    ":" + config.Port,
		Handler: handler,
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on port %s", config.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := HealthResponse{
		Status:    "ok",
		Database:  "disconnected",
		Tailscale: "unknown",
	}

	// Check database
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err == nil {
		health.Database = "connected"
	}

	// Check Tailscale status
	status, err := s.client.Status(r.Context())
	if err == nil && status != nil {
		if status.BackendState == "Running" {
			health.Tailscale = "connected"
		} else {
			health.Tailscale = string(status.BackendState)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

func (s *Server) userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userInfo := UserInfo{Connected: false}

	// Get Tailscale WHOIS information
	whois, err := s.tailscaleWhois(r.Context(), r)
	if err != nil {
		// Only set error if it's not a "daemon not available" error
		// When running without Tailscale or in Docker, we just show not connected
		log.Printf("Tailscale lookup warning: %v", err)
		userInfo.Error = "Tailscale not available"
	}

	if whois != nil {
		userInfo.Connected = true
		userInfo.LoginName = whois.LoginName
		userInfo.DisplayName = whois.DisplayName

		// Get first initial
		if userInfo.DisplayName != "" {
			userInfo.FirstInitial = string(userInfo.DisplayName[0])
		} else if userInfo.LoginName != "" {
			userInfo.FirstInitial = string(userInfo.LoginName[0])
		}
	}

	json.NewEncoder(w).Encode(userInfo)
}

func (s *Server) productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query all columns from products table dynamically
	query := `
		SELECT *
		FROM products
		ORDER BY created_at DESC
		LIMIT 100
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to query database: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Get column names dynamically
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to get columns: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Create a slice to hold the results as maps
	var products []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to scan row: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// Create a map for this row
		product := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert byte arrays to strings
			if b, ok := val.([]byte); ok {
				product[col] = string(b)
			} else if t, ok := val.(time.Time); ok {
				// Format time values as RFC3339
				product[col] = t.Format(time.RFC3339)
			} else {
				product[col] = val
			}
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Error iterating rows: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}

func (s *Server) tailscaleWhois(ctx context.Context, r *http.Request) (*WhoIsData, error) {
	var u *WhoIsData

	// First check for Tailscale identity headers (works in Docker/behind Tailscale Serve)
	// https://tailscale.com/kb/1312/serve#identity-headers
	if r.Header.Get("Tailscale-User-Login") != "" {
		u = &WhoIsData{
			LoginName:   r.Header.Get("Tailscale-User-Login"),
			DisplayName: r.Header.Get("Tailscale-User-Name"),
		}
		return u, nil
	}

	// Try to get WHOIS info from local Tailscale client (works in tsnet mode)
	whois, err := s.client.WhoIs(ctx, r.RemoteAddr)

	if err != nil {
		// Provide helpful error message based on mode
		if s.tsnetMode {
			return nil, fmt.Errorf("failed to identify user via tsnet: %w", err)
		}
		// Not in tsnet mode and no headers - user needs to access via Tailscale
		return nil, fmt.Errorf("not accessed via Tailscale - use 'tailscale serve' or set TS_AUTHKEY to run in tsnet mode")
	}

	// Extract user info from WhoIs response
	if whois.Node.IsTagged() {
		return nil, fmt.Errorf("tagged nodes do not have a user identity")
	} else if whois.UserProfile == nil || whois.UserProfile.LoginName == "" {
		return nil, fmt.Errorf("failed to identify remote user")
	}

	u = &WhoIsData{
		LoginName:   whois.UserProfile.LoginName,
		DisplayName: whois.UserProfile.DisplayName,
	}

	return u, nil
}
