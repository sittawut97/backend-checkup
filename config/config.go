package config

import (
	"os"
	"strings"
)

type Config struct {
	SupabaseURL         string
	SupabaseAnonKey     string
	SupabaseServiceKey  string
	JWTSecret           string
	Port                string
	Environment         string
	AllowedOrigins      []string
	JWTSecret           string
	SMSMKTKey           string
	SMSMKTSender        string
	SMSMKTURL           string
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
		JWTSecret:          os.Getenv("JWT_SECRET"),
		SMSMKTKey:          os.Getenv("SMSMKT_API_KEY"),
		SMSMKTSender:       os.Getenv("SMSMKT_SENDER"),
		SMSMKTURL:          os.Getenv("SMSMKT_URL"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
