package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/api/v1"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/database"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/health"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/middleware"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/observability"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "gateway",
		Short:   "FullStackArkham Inference Gateway",
		Long:    "Central policy and inference control plane for the FullStackArkham platform",
		Version: version,
		RunE:    run,
	}

	rootCmd.Flags().String("host", "0.0.0.0", "Host to bind to")
	rootCmd.Flags().Int("port", 8080, "Port to bind to")
	rootCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) error {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	logLevel, _ := cmd.Flags().GetString("log-level")

	// Initialize observability
	obsConfig := observability.Config{
		ServiceName: "gateway",
		Version:     version,
		LogLevel:    logLevel,
	}
	obs, err := observability.New(obsConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize observability: %w", err)
	}
	defer obs.Shutdown()

	// Initialize database connection
	dbConfig := database.Config{
		Host:            os.Getenv("DATABASE_HOST"),
		Port:            5432,
		User:            os.Getenv("DATABASE_USER"),
		Password:        os.Getenv("DATABASE_PASSWORD"),
		Database:        os.Getenv("DATABASE_NAME"),
		SSLMode:         os.Getenv("DATABASE_SSL_MODE"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
	
	// Set defaults
	if dbConfig.Host == "" {
		dbConfig.Host = "localhost"
	}
	if dbConfig.User == "" {
		dbConfig.User = "postgres"
	}
	if dbConfig.Password == "" {
		dbConfig.Password = "postgres"
	}
	if dbConfig.Database == "" {
		dbConfig.Database = "fullstackarkham"
	}
	if dbConfig.SSLMode == "" {
		dbConfig.SSLMode = "disable"
	}
	
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		obs.Logger.Error("failed to initialize database", "error", err)
		// Continue without database for development
	} else {
		defer db.Close()
		obs.Logger.Info("database connected")
	}

	// Initialize tenant store
	var tenantStore *database.TenantStore
	if db != nil {
		tenantStore = database.NewTenantStore(db.DB)
	}

	// Initialize Redis connection
	// redis, err := cache.NewRedis()
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize redis: %w", err)
	// }

	// Initialize Arkham security client
	arkhamEnabled := os.Getenv("ARKHAM_ENABLED") != "false"
	if arkhamEnabled {
		arkhamEndpoint := os.Getenv("ARKHAM_ENDPOINT")
		if arkhamEndpoint == "" {
			arkhamEndpoint = "http://localhost:8081"
		}
		middleware.InitArkham(arkhamEndpoint, "dev-api-key")
		log.Println("Arkham security initialized at", arkhamEndpoint)
	}

	// Create router
	r := mux.NewRouter()

	// Health and readiness endpoints
	r.HandleFunc("/health", health.HealthHandler).Methods("GET")
	r.HandleFunc("/ready", health.ReadyHandler).Methods("GET")

	// API v1 routes
	apiV1 := r.PathPrefix("/v1").Subrouter()

	// Add tenant store to context for all routes
	apiV1.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if tenantStore != nil {
				ctx = context.WithValue(ctx, middleware.ContextKey("tenant_store"), tenantStore)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Arkham security middleware (applies to all routes)
	if arkhamEnabled {
		apiV1.Use(middleware.ArkhamMiddleware)
	}

	// Auth middleware
	apiV1.Use(middleware.AuthMiddleware)

	// Rate limiting middleware
	apiV1.Use(middleware.RateLimitMiddleware)

	// Observability middleware
	apiV1.Use(middleware.ObservabilityMiddleware(obs))

	// Inference endpoint - main gateway entry point
	apiV1.HandleFunc("/ai", v1.InferenceHandler).Methods("POST")

	// Tenant management endpoints
	apiV1.HandleFunc("/tenants", v1.ListTenantsHandler).Methods("GET")
	apiV1.HandleFunc("/tenants/{id}", v1.GetTenantHandler).Methods("GET")

	// Usage and billing endpoints
	apiV1.HandleFunc("/usage", v1.GetUsageHandler).Methods("GET")
	apiV1.HandleFunc("/billing", v1.GetBillingHandler).Methods("GET")

	// Create server
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		obs.Logger.Info("starting gateway", "addr", addr, "version", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			obs.Logger.Error("server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	obs.Logger.Info("shutting down gateway...")
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	obs.Logger.Info("gateway stopped gracefully")
	return nil
}
