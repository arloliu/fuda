// Example: Validation with go-playground/validator
//
// This example demonstrates using the validate tag for
// configuration validation.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/arloliu/fuda"
)

// Config with validation rules
type Config struct {
	// Required fields
	AppName string `yaml:"app_name" validate:"required"`

	// Numeric constraints
	Port        int `yaml:"port" validate:"required,min=1,max=65535"`
	MaxWorkers  int `yaml:"max_workers" validate:"required,min=1,max=100"`
	MaxBodySize int `yaml:"max_body_size" validate:"min=0,max=104857600"` // 100MB max

	// String constraints
	Environment string `yaml:"environment" validate:"required,oneof=dev staging prod"`
	LogLevel    string `yaml:"log_level" validate:"required,oneof=debug info warn error"`

	// URL validation
	Endpoint string `yaml:"endpoint" validate:"required,url"`

	// Email validation
	AdminEmail string `yaml:"admin_email" validate:"required,email"`

	// Nested validation
	Database DatabaseConfig `yaml:"database" validate:"required"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" validate:"required,hostname|ip"`
	Port     int    `yaml:"port" validate:"required,min=1,max=65535"`
	Database string `yaml:"database" validate:"required,min=1,max=63"`
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required,min=8"`
}

func main() {
	// Try loading valid config
	fmt.Println("=== Loading valid configuration ===")
	if err := loadConfig("config_valid.yaml"); err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Configuration is valid!")
	}

	// Try loading invalid config
	fmt.Println("\n=== Loading invalid configuration ===")
	if err := loadConfig("config_invalid.yaml"); err != nil {
		handleValidationError(err)
	}
}

func loadConfig(path string) error {
	var cfg Config
	return fuda.LoadFile(path, &cfg)
}

func handleValidationError(err error) {
	var validationErr *fuda.ValidationError
	if errors.As(err, &validationErr) {
		fmt.Println("Validation failed:")
		for _, fieldErr := range validationErr.Errors {
			fmt.Printf("  ✗ %v\n", fieldErr)
		}
	} else {
		fmt.Printf("Other error: %v\n", err)
	}
}
