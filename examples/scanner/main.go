// Example: Custom types with Scanner interface
//
// This example demonstrates implementing the Scanner interface
// for custom type conversion from default tag values.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/arloliu/fuda"
)

// LogLevel is a custom type with string parsing
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// Scan implements fuda.Scanner for parsing default tag values
func (l *LogLevel) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}

	switch strings.ToLower(s) {
	case "debug":
		*l = LogDebug
	case "info":
		*l = LogInfo
	case "warn", "warning":
		*l = LogWarn
	case "error":
		*l = LogError
	default:
		return fmt.Errorf("unknown log level: %s", s)
	}
	return nil
}

func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarn:
		return "WARN"
	case LogError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// DatabaseDriver is another custom enum type
type DatabaseDriver string

const (
	DriverPostgres DatabaseDriver = "postgres"
	DriverMySQL    DatabaseDriver = "mysql"
	DriverSQLite   DatabaseDriver = "sqlite"
)

func (d *DatabaseDriver) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}

	switch strings.ToLower(s) {
	case "postgres", "postgresql", "pg":
		*d = DriverPostgres
	case "mysql", "mariadb":
		*d = DriverMySQL
	case "sqlite", "sqlite3":
		*d = DriverSQLite
	default:
		return fmt.Errorf("unknown driver: %s", s)
	}
	return nil
}

// Config using custom Scanner types
type Config struct {
	AppName  string         `yaml:"app_name" default:"scanner-example"`
	LogLevel LogLevel       `yaml:"log_level" default:"info"`
	Driver   DatabaseDriver `yaml:"driver" default:"postgres"`
}

func main() {
	var cfg Config
	if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App: %s\n", cfg.AppName)
	fmt.Printf("Log Level: %s (value: %d)\n", cfg.LogLevel, cfg.LogLevel)
	fmt.Printf("Driver: %s\n", cfg.Driver)
}
