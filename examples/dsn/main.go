// Example: DSN (Data Source Name) composition
//
// This example demonstrates using the dsn tag to compose connection strings
// from multiple configuration sources including fields, secrets, and env vars.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/arloliu/fuda"
)

// Config demonstrates various DSN composition patterns
type Config struct {
	// === Pattern 1: DSN from Field References ===
	// Fields are populated from YAML, env, and defaults first,
	// then DSN is composed from those fields
	DBHost     string `yaml:"db_host" default:"localhost"`
	DBPort     int    `yaml:"db_port" default:"5432"`
	DBName     string `yaml:"db_name" default:"myapp"`
	DBUser     string `yaml:"db_user" env:"DB_USER"`
	DBPassword string `yaml:"db_password" env:"DB_PASSWORD"`

	// Compose from fields above using ${.FieldName} syntax
	PostgresDSN string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:${.DBPort}/${.DBName}"`

	// === Pattern 2: DSN with Inline Environment Variables ===
	// Use ${env:KEY} to read env vars directly in the DSN template
	RedisHost string `yaml:"redis_host" default:"localhost"`
	RedisPort int    `yaml:"redis_port" default:"6379"`

	// Mix field references with inline env var using ${env:KEY}
	RedisDSN string `dsn:"redis://:${env:REDIS_PASSWORD}@${.RedisHost}:${.RedisPort}/0"`

	// === Pattern 3: DSN with Inline File Reference ===
	// Use ${ref:uri} to resolve secrets from files inline
	MongoHost     string `yaml:"mongo_host" default:"localhost"`
	MongoPassword string // Will be set dynamically with absolute path in DSN
}

func main() {
	// Set up example environment variables
	os.Setenv("DB_USER", "postgres_admin")
	os.Setenv("DB_PASSWORD", "postgres_secret")
	os.Setenv("REDIS_PASSWORD", "redis_secret")
	defer func() {
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("REDIS_PASSWORD")
	}()

	loader, err := fuda.New().
		FromFile("config.yaml").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== DSN Composition Example ===")
	fmt.Println()

	// Pattern 1: Field References
	fmt.Println("--- Pattern 1: DSN from Field References ---")
	fmt.Println("Config struct uses: `dsn:\"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:${.DBPort}/${.DBName}\"`")
	fmt.Printf("  DB Host:       %s (from yaml/default)\n", cfg.DBHost)
	fmt.Printf("  DB User:       %s (from env:DB_USER)\n", cfg.DBUser)
	fmt.Printf("  PostgreSQL DSN: %s\n", maskPassword(cfg.PostgresDSN))
	fmt.Println()

	// Pattern 2: Inline Env Vars
	fmt.Println("--- Pattern 2: DSN with Inline ${env:KEY} ---")
	fmt.Println("Config struct uses: `dsn:\"redis://:${env:REDIS_PASSWORD}@${.RedisHost}:${.RedisPort}/0\"`")
	fmt.Printf("  Redis Host:    %s\n", cfg.RedisHost)
	fmt.Printf("  Redis DSN:     %s\n", maskPassword(cfg.RedisDSN))
	fmt.Println()

	// Pattern 3: Inline File Reference (demonstrate with dynamic struct)
	fmt.Println("--- Pattern 3: DSN with Inline ${ref:file://...} ---")
	demonstrateFileRef()
}

// demonstrateFileRef shows how to use ${ref:file://...} with absolute paths
func demonstrateFileRef() {
	// Create a temporary secret file
	tmpDir, _ := os.MkdirTemp("", "fuda-example")
	defer os.RemoveAll(tmpDir)

	secretFile := filepath.Join(tmpDir, "mongo_password.txt")
	os.WriteFile(secretFile, []byte("mongo_secret_123"), 0o600)

	// For file:// refs, use absolute paths: file:///absolute/path
	fmt.Printf("  Secret file:   %s\n", secretFile)
	fmt.Println("  DSN template:  `dsn:\"mongodb://admin:${ref:file:///path/to/secret}@host:27017/db\"`")
	fmt.Printf("  Would resolve: mongodb://admin:mongo_secret_123@localhost:27017/db\n")
}

// maskPassword replaces passwords in DSN for safe display
func maskPassword(dsn string) string {
	// Simple masking for demo - find :password@ pattern
	inPassword := false
	result := make([]byte, 0, len(dsn))

	for i := 0; i < len(dsn); i++ {
		if !inPassword && i > 0 && dsn[i-1] == ':' && (i < 3 || dsn[i-3:i] != "://") {
			// Start of password (after : but not after ://)
			inPassword = true
			result = append(result, '*', '*', '*', '*')
		}
		if inPassword && dsn[i] == '@' {
			inPassword = false
		}
		if !inPassword {
			result = append(result, dsn[i])
		}
	}

	return string(result)
}
