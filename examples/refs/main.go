// Example: External references with ref and refFrom tags
//
// This example demonstrates loading secrets and configuration from
// external files using the ref and refFrom tags.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arloliu/fuda"
)

// Config with external references
type Config struct {
	AppName string `yaml:"app_name" default:"my-app"`

	// Static reference - load from fixed path
	// The ref tag loads content from file:// URIs
	APIKey string `ref:"file://secrets/api_key.txt"`

	// Dynamic reference - path comes from another field
	// refFrom reads the URI from the specified field
	DatabasePasswordPath string `yaml:"db_password_path"`
	DatabasePassword     string `refFrom:"DatabasePasswordPath"`

	// Nested config with refs
	TLS TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	CertPath string `yaml:"cert_path"`
	KeyPath  string `yaml:"key_path"`
	Cert     string `refFrom:"CertPath"`
	Key      string `refFrom:"KeyPath"`
}

func main() {
	// Create sample secret files
	setupSecrets()
	defer cleanupSecrets()

	loader, err := fuda.New().
		FromFile("config.yaml").
		WithTimeout(5 * time.Second). // Timeout for HTTP refs
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App: %s\n", cfg.AppName)
	fmt.Printf("API Key: %s\n", cfg.APIKey)
	fmt.Printf("DB Password: %s\n", cfg.DatabasePassword)
	fmt.Printf("TLS Cert loaded: %d bytes\n", len(cfg.TLS.Cert))
	fmt.Printf("TLS Key loaded: %d bytes\n", len(cfg.TLS.Key))
}

func setupSecrets() {
	os.MkdirAll("secrets", 0o755)
	os.WriteFile("secrets/api_key.txt", []byte("sk-1234567890abcdef"), 0o600)
	os.WriteFile("secrets/db_password.txt", []byte("super-secret-password"), 0o600)
	os.WriteFile("secrets/server.crt", []byte("-----BEGIN CERTIFICATE-----\n...cert data...\n-----END CERTIFICATE-----"), 0o600)
	os.WriteFile("secrets/server.key", []byte("-----BEGIN PRIVATE KEY-----\n...key data...\n-----END PRIVATE KEY-----"), 0o600)
}

func cleanupSecrets() {
	os.RemoveAll("secrets")
}
