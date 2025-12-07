package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/K-H-Tech/apm/internal/contract"
	"github.com/K-H-Tech/apm/internal/handlers"
	"github.com/K-H-Tech/apm/internal/middleware"
	"github.com/K-H-Tech/apm/internal/repository"
	"github.com/K-H-Tech/apm/internal/service"
	"github.com/K-H-Tech/apm/pkg/atlassian"
	"github.com/K-H-Tech/apm/pkg/llm"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Config holds application configuration
type Config struct {
	// Server
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CORSOrigins  []string

	// Database
	DatabaseURL string

	// LLM
	OpenAIKey       string
	AnthropicKey    string
	DefaultProvider string
	DefaultModel    string

	// Atlassian OAuth
	AtlassianClientID     string
	AtlassianClientSecret string
	AtlassianRedirectURL  string

	// Security
	EncryptionKey string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		// Server
		Port:         getEnv("PORT", "8080"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
		CORSOrigins:  getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/apm?sslmode=disable"),

		// LLM
		OpenAIKey:       getEnv("OPENAI_API_KEY", ""),
		AnthropicKey:    getEnv("ANTHROPIC_API_KEY", ""),
		DefaultProvider: getEnv("LLM_DEFAULT_PROVIDER", "openai"),
		DefaultModel:    getEnv("LLM_DEFAULT_MODEL", "gpt-4"),

		// Atlassian OAuth
		AtlassianClientID:     getEnv("ATLASSIAN_CLIENT_ID", ""),
		AtlassianClientSecret: getEnv("ATLASSIAN_CLIENT_SECRET", ""),
		AtlassianRedirectURL:  getEnv("ATLASSIAN_REDIRECT_URL", "http://localhost:8080/auth/atlassian/callback"),

		// Security
		EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
	}
}

func main() {
	// Load configuration
	cfg := LoadConfig()

	// Validate required configuration
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	prdRepo := repository.NewPRDRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	priorityRepo := repository.NewPriorityRepository(db)

	// Initialize LLM provider
	llmProvider := initLLMProvider(cfg)

	// Initialize services
	aiService := service.NewAIService(llmProvider)
	prdService := service.NewPRDService(prdRepo, templateRepo, aiService)
	priorityService := service.NewPriorityService(priorityRepo)

	// Initialize OAuth client
	oauthClient := atlassian.NewOAuthClient(atlassian.OAuthConfig{
		ClientID:     cfg.AtlassianClientID,
		ClientSecret: cfg.AtlassianClientSecret,
		RedirectURL:  cfg.AtlassianRedirectURL,
		Scopes:       atlassian.DefaultScopes(),
	})

	// Initialize encrypter (placeholder - implement proper encryption)
	encrypter := &simpleEncrypter{key: cfg.EncryptionKey}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(handlers.AuthHandlerConfig{
		OAuthClient: oauthClient,
		UserRepo:    userRepo,
		OrgRepo:     orgRepo,
		Encrypter:   encrypter,
	})

	prdHandler := handlers.NewPRDHandler(prdService)
	priorityHandler := handlers.NewPriorityHandler(priorityService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Organization-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// Auth routes (public)
	authHandler.RegisterRoutes(router.Group(""))

	// API routes
	api := router.Group("/api/v1")

	// Add auth middleware
	api.Use(middleware.AuthMiddleware(middleware.AuthConfig{
		UserRepo: userRepo,
		OrgRepo:  orgRepo,
		SkipPaths: []string{
			"/auth/atlassian",
			"/auth/atlassian/callback",
			"/health",
		},
	}))

	// Add rate limiting
	api.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimitConfig()))

	// Register routes
	prdHandler.RegisterRoutes(api)
	priorityHandler.RegisterRoutes(api)

	// AI routes with stricter rate limiting
	aiRoutes := api.Group("/ai")
	aiRoutes.Use(middleware.AIRateLimitMiddleware())
	// AI routes are already registered via prdHandler.RegisterRoutes

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// initLLMProvider initializes the LLM provider based on configuration
func initLLMProvider(cfg *Config) contract.LLMProvider {
	// Build LLM config
	llmConfig := &llm.Config{
		OpenAIAPIKey:    cfg.OpenAIKey,
		AnthropicAPIKey: cfg.AnthropicKey,
		MaxTokens:       4096,
		Temperature:     0.7,
		Timeout:         60 * time.Second,
		MaxRetries:      3,
	}

	// Set default model based on available providers
	if cfg.DefaultModel != "" {
		llmConfig.OpenAIModel = cfg.DefaultModel
		llmConfig.AnthropicModel = cfg.DefaultModel
	}

	// Set default provider
	if cfg.DefaultProvider == "anthropic" && cfg.AnthropicKey != "" {
		llmConfig.DefaultProvider = llm.ProviderAnthropic
	} else if cfg.OpenAIKey != "" {
		llmConfig.DefaultProvider = llm.ProviderOpenAI
	} else if cfg.AnthropicKey != "" {
		llmConfig.DefaultProvider = llm.ProviderAnthropic
	}

	var primary, fallback contract.LLMProvider
	var err error

	// Create primary provider
	if cfg.OpenAIKey != "" {
		primary, err = llm.NewOpenAIProvider(llmConfig)
		if err != nil {
			log.Printf("Warning: Failed to create OpenAI provider: %v", err)
		}
	}

	// Create fallback provider
	if cfg.AnthropicKey != "" && primary != nil {
		fallback, err = llm.NewAnthropicProvider(llmConfig)
		if err != nil {
			log.Printf("Warning: Failed to create Anthropic provider: %v", err)
		}
	} else if cfg.AnthropicKey != "" {
		// Anthropic is primary if OpenAI not configured
		primary, err = llm.NewAnthropicProvider(llmConfig)
		if err != nil {
			log.Printf("Warning: Failed to create Anthropic provider: %v", err)
		}
	}

	if primary == nil {
		log.Println("Warning: No LLM providers configured")
		return nil
	}

	// Use multi-provider with fallback if available
	if fallback != nil {
		return llm.NewMultiProvider(primary, fallback, 3)
	}

	return primary
}

// validateConfig validates required configuration
func validateConfig(cfg *Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.AtlassianClientID == "" || cfg.AtlassianClientSecret == "" {
		log.Println("Warning: Atlassian OAuth not configured")
	}

	if cfg.OpenAIKey == "" && cfg.AnthropicKey == "" {
		log.Println("Warning: No LLM API keys configured - AI features will be disabled")
	}

	return nil
}

// Helper functions for loading config from environment
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		var result []string
		for _, v := range splitAndTrim(value, ",") {
			if v != "" {
				result = append(result, v)
			}
		}
		return result
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i < len(s)-len(sep)+1 && s[i:i+len(sep)] == sep {
			result = append(result, trim(s[start:i]))
			start = i + len(sep)
		}
	}
	result = append(result, trim(s[start:]))
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

// simpleEncrypter is a placeholder encrypter
// In production, use proper AES encryption from pkg/crypto
type simpleEncrypter struct {
	key string
}

func (e *simpleEncrypter) Encrypt(plaintext string) (string, error) {
	// Placeholder - implement proper encryption
	return plaintext, nil
}

func (e *simpleEncrypter) Decrypt(ciphertext string) (string, error) {
	// Placeholder - implement proper decryption
	return ciphertext, nil
}
