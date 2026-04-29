// Package main implements the Papabase vertical service
// Papabase is a family studio suite built on top of stack-arkham core
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

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/cobra"


	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/api"
	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	version = "0.1.0"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "papabase",
		Short:   "Papabase Family Studio Suite",
		Long:    "Business operations platform with AI-powered website generation (Dad AI)",
		Version: version,
		RunE:    run,
	}

	rootCmd.Flags().String("host", "0.0.0.0", "Host to bind to")
	rootCmd.Flags().Int("port", 8087, "Port to bind to")
	rootCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().String("gateway-url", "https://gateway.redmushroom-ed7ab088.westus.azurecontainerapps.io", "Stack-arkham gateway URL")
	rootCmd.Flags().String("db-host", os.Getenv("DATABASE_HOST"), "Database host")
	rootCmd.Flags().String("db-user", os.Getenv("DATABASE_USER"), "Database user")
	rootCmd.Flags().String("db-password", os.Getenv("DATABASE_PASSWORD"), "Database password")
	rootCmd.Flags().String("db-name", os.Getenv("DATABASE_NAME"), "Database name")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) error {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	logLevel, _ := cmd.Flags().GetString("log-level")
	gatewayURL, _ := cmd.Flags().GetString("gateway-url")

	dbHost, _ := cmd.Flags().GetString("db-host")
	dbUser, _ := cmd.Flags().GetString("db-user")
	dbPass, _ := cmd.Flags().GetString("db-password")
	dbName, _ := cmd.Flags().GetString("db-name")

	// Set log level
	if logLevel == "debug" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Initialize CRM store
	var crmStore api.CRMStore
	if dbHost != "" {
		if dbUser == "" {
			dbUser = "postgres"
		}
		if dbPass == "" {
			dbPass = "postgres"
		}
		if dbName == "" {
			dbName = "fullstackarkham"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=require", dbHost, dbUser, dbPass, dbName)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		
		// Auto-migrate models
		if err := db.AutoMigrate(&api.LeadModel{}, &api.TaskModel{}); err != nil {
			log.Printf("Warning: auto-migration failed: %v", err)
		}
		
		crmStore = api.NewSQLCRMStore(db)
		log.Printf("SQL CRM store initialized (host: %s, db: %s)", dbHost, dbName)
	} else {
		crmStore = api.NewInMemoryCRMStore()
		log.Println("In-memory CRM store initialized (development mode)")
	}

	// Initialize Dad AI agent
	ctx := context.Background()
	dadAI, err := agents.NewDadAIAgent(ctx, gatewayURL)

	if err != nil {
		return fmt.Errorf("failed to initialize Dad AI agent: %w", err)
	}
	log.Println("Dad AI agent initialized")

	// Initialize Lead Scorer agent
	leadScorer := agents.NewLeadScorer(gatewayURL)
	log.Println("Lead Scorer agent initialized")

	// Initialize Task Agent
	taskAgent := agents.NewTaskAgent(gatewayURL)
	log.Println("Task Agent initialized")

	// Initialize Content Agent
	contentAgent := agents.NewContentAgent(gatewayURL)
	log.Println("Content Agent initialized")

	// Initialize Proposal Agent
	proposalAgent := agents.NewProposalAgent(gatewayURL)
	log.Println("Proposal Agent initialized")

	// Initialize GTM Agent
	gtmAgent := agents.NewGTMAgent(gatewayURL)
	log.Println("GTM Agent initialized")

	// Initialize Azure Deployment Bot
	azureDeploymentBot := agents.NewAzureDeploymentBot(gatewayURL)
	log.Println("Azure Deployment Bot initialized")

	// Initialize Debug Agent
	debugAgent := agents.NewDebugAgent(gatewayURL)
	log.Println("Debug Agent initialized")

	// Initialize Code Agent
	codeAgent := agents.NewCodeAgent(gatewayURL)
	log.Println("Code Agent initialized")

	// Create router
	r := mux.NewRouter()



	// Health endpoints
	r.HandleFunc("/health", api.HealthHandler).Methods("GET")
	r.HandleFunc("/ready", api.ReadyHandler).Methods("GET")

	// API v1 routes
	apiV1 := r.PathPrefix("/api/v1").Subrouter()

	// Add context values
	apiV1.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, api.ContextKey("dad_ai"), dadAI)
			ctx = context.WithValue(ctx, api.ContextKey("lead_scorer"), leadScorer)
			ctx = context.WithValue(ctx, api.ContextKey("task_agent"), taskAgent)
			ctx = context.WithValue(ctx, api.ContextKey("content_agent"), contentAgent)
			ctx = context.WithValue(ctx, api.ContextKey("proposal_agent"), proposalAgent)
			ctx = context.WithValue(ctx, api.ContextKey("gtm_agent"), gtmAgent)
			ctx = context.WithValue(ctx, api.ContextKey("azure_deployment_bot"), azureDeploymentBot)
			ctx = context.WithValue(ctx, api.ContextKey("debug_agent"), debugAgent)
			ctx = context.WithValue(ctx, api.ContextKey("code_agent"), codeAgent)
			ctx = context.WithValue(ctx, api.ContextKey("crm_store"), crmStore)

			ctx = context.WithValue(ctx, api.ContextKey("gateway_url"), gatewayURL)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// CRM endpoints
	apiV1.HandleFunc("/leads", api.CreateLeadHandler).Methods("POST")
	apiV1.HandleFunc("/leads", api.ListLeadsHandler).Methods("GET")
	apiV1.HandleFunc("/leads/{id}", api.GetLeadHandler).Methods("GET")
	apiV1.HandleFunc("/leads/{id}", api.UpdateLeadHandler).Methods("PUT")
	apiV1.HandleFunc("/leads/{id}", api.DeleteLeadHandler).Methods("DELETE")

	// Task endpoints
	apiV1.HandleFunc("/tasks", api.CreateTaskHandler).Methods("POST")
	apiV1.HandleFunc("/tasks", api.ListTasksHandler).Methods("GET")
	apiV1.HandleFunc("/tasks/{id}", api.GetTaskHandler).Methods("GET")
	apiV1.HandleFunc("/tasks/{id}", api.UpdateTaskHandler).Methods("PUT")
	apiV1.HandleFunc("/tasks/{id}", api.DeleteTaskHandler).Methods("DELETE")

	// Dad AI website generation endpoints
	apiV1.HandleFunc("/ai/generate", api.GenerateWebsiteHandler).Methods("POST")
	apiV1.HandleFunc("/ai/templates", api.ListTemplatesHandler).Methods("GET")
	apiV1.HandleFunc("/ai/projects/{id}", api.GetProjectHandler).Methods("GET")
	apiV1.HandleFunc("/ai/projects", api.ListProjectsHandler).Methods("GET")

	// AI Lead Scoring endpoints
	apiV1.HandleFunc("/ai/leads/score", api.ScoreLeadHandler).Methods("POST")
	apiV1.HandleFunc("/ai/leads/score/batch", api.ScoreLeadsBatchHandler).Methods("POST")
	apiV1.HandleFunc("/ai/leads/{id}/insights", api.GetLeadInsightsHandler).Methods("GET")

	// AI Task Agent endpoints
	apiV1.HandleFunc("/ai/tasks/generate", api.GenerateTaskHandler).Methods("POST")
	apiV1.HandleFunc("/ai/tasks/breakdown", api.BreakdownTaskHandler).Methods("POST")

	// AI Content Generator endpoints
	apiV1.HandleFunc("/ai/content/business-description", api.GenerateBusinessDescriptionHandler).Methods("POST")
	apiV1.HandleFunc("/ai/content/seo-meta", api.GenerateSEOMetaHandler).Methods("POST")
	apiV1.HandleFunc("/ai/content/blog", api.GenerateBlogPostHandler).Methods("POST")
	apiV1.HandleFunc("/ai/content/social", api.GenerateSocialMediaHandler).Methods("POST")
	apiV1.HandleFunc("/ai/content/faq", api.GenerateFAQHandler).Methods("POST")

	// AI Proposal Builder endpoints
	apiV1.HandleFunc("/ai/proposals/generate", api.GenerateProposalHandler).Methods("POST")
	apiV1.HandleFunc("/ai/proposals/quote", api.GenerateQuoteHandler).Methods("POST")
	apiV1.HandleFunc("/ai/proposals/scope", api.GenerateScopeHandler).Methods("POST")
	apiV1.HandleFunc("/ai/proposals/pricing-tiers", api.GeneratePricingTiersHandler).Methods("POST")

	// GTM Agent endpoints
	apiV1.HandleFunc("/ai/gtm/strategy", api.GenerateGTMStrategyHandler).Methods("POST")
	apiV1.HandleFunc("/ai/gtm/outbound", api.GenerateGTMOutboundHandler).Methods("POST")

	// Azure Deployment Bot endpoints
	apiV1.HandleFunc("/ai/deploy/azure", api.PlanAzureDeploymentHandler).Methods("POST")

	// Debug Agent endpoints
	apiV1.HandleFunc("/ai/debug", api.DebugErrorHandler).Methods("POST")

	// Code Agent endpoints
	apiV1.HandleFunc("/ai/code/generate", api.GenerateCodeHandler).Methods("POST")
	apiV1.HandleFunc("/ai/code/refactor", api.RefactorCodeHandler).Methods("POST")

	// Pricing & Billing

	apiV1.HandleFunc("/pricing/plans", api.ListPlansHandler).Methods("GET")
	apiV1.HandleFunc("/billing/usage", api.GetUsageHandler).Methods("GET")
	apiV1.HandleFunc("/billing/invoices", api.ListInvoicesHandler).Methods("GET")

	// Create server
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://fsai.pro", "https://www.fsai.pro", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	})

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      c.Handler(r),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}


	// Start server in goroutine
	go func() {
		log.Printf("Starting Papabase service on %s (version %s)", addr, version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Shutting down Papabase service...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Papabase service stopped gracefully")
	return nil
}
