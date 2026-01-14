package config

import (
	"log"

	supa "github.com/supabase-community/supabase-go"
)

func NewSupabaseClient(cfg *Config) *supa.Client {
	client, err := supa.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, nil)
	if err != nil {
		log.Fatalf("Failed to create Supabase client: %v", err)
	}
	return client
}
