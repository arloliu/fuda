// Example: Dotenv file loading
//
// This example demonstrates loading environment variables from .env files
// before processing configuration, enabling environment-specific overlays.
package main

import (
	"fmt"
	"log"

	"github.com/arloliu/fuda"
)

// Config demonstrates dotenv integration with fuda.
type Config struct {
	// Database configuration from .env file
	DBHost     string `yaml:"db_host" env:"DB_HOST" default:"localhost"`
	DBPort     int    `yaml:"db_port" env:"DB_PORT" default:"5432"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME" default:"myapp"`

	// App settings from .env or config file
	AppEnv   string `yaml:"app_env" env:"APP_ENV" default:"development"`
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" default:"info"`
}

func main() {
	// Example 1: Load from single .env file
	fmt.Println("=== Loading with .env file ===")

	loader, err := fuda.New().
		FromFile("config.yaml").
		WithDotEnv(".env"). // Load .env before processing
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database: %s@%s:%d/%s\n", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	fmt.Printf("Environment: %s, LogLevel: %s\n", cfg.AppEnv, cfg.LogLevel)

	// Example 2: Using overlay pattern for environment-specific config
	fmt.Println("\n=== Using overlay pattern ===")

	loaderOverlay, err := fuda.New().
		FromFile("config.yaml").
		// Load base .env, then .env.local for local overrides
		WithDotEnvFiles([]string{".env", ".env.local"}).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg2 Config
	if err := loaderOverlay.Load(&cfg2); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database: %s@%s:%d/%s\n", cfg2.DBUser, cfg2.DBHost, cfg2.DBPort, cfg2.DBName)

	// Example 3: Override mode (dotenv values override existing env vars)
	fmt.Println("\n=== Using override mode ===")

	loaderOverride, err := fuda.New().
		FromFile("config.yaml").
		WithDotEnv(".env.production", fuda.DotEnvOverride()).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg3 Config
	if err := loaderOverride.Load(&cfg3); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Environment (overridden): %s\n", cfg3.AppEnv)
}
