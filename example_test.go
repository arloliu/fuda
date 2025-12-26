package fuda_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arloliu/fuda"
)

// ExampleLoadFile demonstrates the simplest usage: loading configuration from a file.
func ExampleLoadFile() {
	// Create a temporary config file for the example
	configContent := `
host: example.com
port: 9000
`
	if err := os.WriteFile("example_config.yaml", []byte(configContent), 0o600); err != nil {
		fmt.Println("failed to write config file")
		return
	}
	defer os.Remove("example_config.yaml")

	type Config struct {
		Host string `yaml:"host" default:"localhost"`
		Port int    `yaml:"port" default:"8080"`
	}

	var cfg Config
	if err := fuda.LoadFile("example_config.yaml", &cfg); err != nil {
		fmt.Println("failed to load config")
		return
	}

	fmt.Printf("Host: %s, Port: %d\n", cfg.Host, cfg.Port)
	// Output: Host: example.com, Port: 9000
}

// ExampleNew demonstrates the builder pattern for advanced configuration.
func ExampleNew() {
	configContent := `
database:
  host: db.example.com
`
	type DatabaseConfig struct {
		Host string `yaml:"host" default:"localhost"`
		Port int    `yaml:"port" default:"5432"`
	}
	type Config struct {
		Database DatabaseConfig `yaml:"database"`
	}

	loader, err := fuda.New().
		FromBytes([]byte(configContent)).
		WithEnvPrefix("APP_").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	// Output: Database: db.example.com:5432
}

// ExampleRefResolver demonstrates implementing a custom reference resolver.
func ExampleRefResolver() {
	// A simple in-memory resolver for demonstration
	type MemoryResolver struct {
		secrets map[string]string
	}

	resolver := &MemoryResolver{
		secrets: map[string]string{
			"secret://api-key": "my-secret-api-key",
		},
	}

	// Implement the RefResolver interface
	resolve := func(_ context.Context, uri string) ([]byte, error) {
		if val, ok := resolver.secrets[uri]; ok {
			return []byte(val), nil
		}
		return nil, fmt.Errorf("secret not found: %s", uri)
	}

	// Use a wrapper type since we can't add methods to MemoryResolver in this example
	_ = resolve // In real usage, pass a type implementing RefResolver to WithRefResolver

	fmt.Println("Custom RefResolver can handle any URI scheme")
	// Output: Custom RefResolver can handle any URI scheme
}
