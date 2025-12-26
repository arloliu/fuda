// Example: Template processing
//
// This example demonstrates using Go templates to dynamically
// generate configuration values before YAML parsing.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/arloliu/fuda"
)

// TemplateData is passed to the template engine
type TemplateData struct {
	Environment string
	Region      string
	Version     string
}

// Config is the final parsed configuration
type Config struct {
	AppName     string `yaml:"app_name"`
	Environment string `yaml:"environment"`
	Endpoint    string `yaml:"endpoint"`
	LogLevel    string `yaml:"log_level"`
	Replicas    int    `yaml:"replicas"`
}

func main() {
	// Template data - could come from environment, flags, etc.
	data := TemplateData{
		Environment: getEnvOrDefault("ENVIRONMENT", "dev"),
		Region:      getEnvOrDefault("REGION", "us-west-2"),
		Version:     "1.2.3",
	}

	fmt.Printf("Template Data: env=%s, region=%s, version=%s\n\n", data.Environment, data.Region, data.Version)

	loader, err := fuda.New().
		FromFile("config.yaml").
		WithTemplate(data).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Rendered Configuration ===")
	fmt.Printf("App Name:    %s\n", cfg.AppName)
	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Endpoint:    %s\n", cfg.Endpoint)
	fmt.Printf("Log Level:   %s\n", cfg.LogLevel)
	fmt.Printf("Replicas:    %d\n", cfg.Replicas)
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
