package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/routes"
	"github.com/sittawut/backend-appointment/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize Supabase client
	supabaseClient := config.NewSupabaseClient(cfg)

	// Initialize SMS client - using SMSMKT (SMS2PRO pending sender name activation)
	smsClient := &services.SMSMKTClient{
		APIKey:     cfg.SMSMKTKey,
		SecretKey:  cfg.SMSMKTSecretKey,
		ProjectKey: cfg.SMSMKTProjectKey,
		URL:        cfg.SMSMKTURL,
	}

	// // SMS2PRO client (commented - waiting for sender name activation)
	// smsClient := services.NewSMS2ProClient(cfg.SMS2ProAPIKey)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Setup CORS middleware
	router.Use(config.CORSMiddleware(cfg))

	// Setup routes
	routes.SetupRoutes(router, supabaseClient, cfg, smsClient)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
