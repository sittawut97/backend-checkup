package config

import (
	"os"
	"strings"
)

type Config struct {
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
	JWTSecret          string
	Port               string
	Environment        string
	AllowedOrigins     []string
	SMSMKTKey          string
	SMSMKTSecretKey    string
	SMSMKTProjectKey   string
	SMSMKTURL          string
	SMS2ProAPIKey      string
	THSMSToken         string
	THSMSBaseURL       string
	THSMSSender        string
	AzureClientID      string
	AzureClientSecret  string
	AzureTenantID      string
	AzureRedirectURI   string
}

func NewConfig() *Config {
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{"http://localhost:3000"}
	if allowedOriginsStr != "" {
		allowedOrigins = strings.Split(allowedOriginsStr, ",")
	}

	return &Config{
		SupabaseURL:        os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:    os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		Port:               getEnvOrDefault("PORT", "8080"),
		Environment:        getEnvOrDefault("ENVIRONMENT", "development"),
		AllowedOrigins:     allowedOrigins,
		SMSMKTKey:          os.Getenv("SMSMKT_API_KEY"),
		SMSMKTSecretKey:    os.Getenv("SMSMKT_SECRET_KEY"),
		SMSMKTProjectKey:   os.Getenv("SMSMKT_PROJECT_KEY"),
		SMSMKTURL:          os.Getenv("SMSMKT_URL"),
		SMS2ProAPIKey:      os.Getenv("SMS2PRO_API_KEY"),
		THSMSToken:         os.Getenv("THSMS_API_TOKEN"),
		THSMSBaseURL:       os.Getenv("THSMS_BASE_URL"),
		THSMSSender:        os.Getenv("THSMS_SENDER"),
		AzureClientID:      os.Getenv("AZURE_CLIENT_ID"),
		AzureClientSecret:  os.Getenv("AZURE_CLIENT_SECRET"),
		AzureTenantID:      os.Getenv("AZURE_TENANT_ID"),
		AzureRedirectURI:   os.Getenv("AZURE_REDIRECT_URI"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
