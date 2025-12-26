// Example: Basic configuration loading with defaults
//
// This example demonstrates loading configuration from a YAML file
// with default values and environment variable overrides.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arloliu/fuda"
)

// Config demonstrates the basic fuda tags for configuration.
type Config struct {
	// App settings with defaults
	AppName string `yaml:"app_name" default:"my-app"`
	Version string `yaml:"version" default:"1.0.0"`

	// Server configuration with env overrides
	Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
	Port int    `yaml:"port" default:"8080" env:"APP_PORT"`

	// Timeout with duration parsing
	Timeout time.Duration `yaml:"timeout" default:"30s"`

	// Feature flags
	Debug   bool `yaml:"debug" default:"false" env:"APP_DEBUG"`
	Verbose bool `yaml:"verbose" default:"false"`
}

func main() {
	// Example 1: Load from file with defaults
	fmt.Println("=== Loading from config.yaml ===")

	var cfg Config
	if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App: %s v%s\n", cfg.AppName, cfg.Version)
	fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("Timeout: %v\n", cfg.Timeout)
	fmt.Printf("Debug: %v\n", cfg.Debug)

	// Example 2: Using the builder pattern with env prefix
	fmt.Println("\n=== Using Builder Pattern ===")

	loader, err := fuda.New().
		FromFile("config.yaml").
		WithEnvPrefix("MYAPP_"). // Reads MYAPP_HOST, MYAPP_PORT, etc.
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg2 Config
	if err := loader.Load(&cfg2); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Host (with prefix): %s\n", cfg2.Host)
}
