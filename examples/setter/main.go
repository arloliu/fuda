// Example: Dynamic defaults with Setter interface
//
// This example demonstrates implementing the Setter interface
// for dynamic default values that depend on runtime data.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arloliu/fuda"
)

// Config with dynamic defaults via Setter interface
type Config struct {
	// Static defaults
	AppName string `yaml:"app_name" default:"setter-example"`

	// Dynamic defaults set by SetDefaults()
	RequestID string    `yaml:"request_id"`
	Hostname  string    `yaml:"hostname"`
	StartTime time.Time `yaml:"start_time"`

	// Nested struct with its own Setter
	Server ServerConfig `yaml:"server"`
}

// SetDefaults is called automatically by fuda when loading
func (c *Config) SetDefaults() {
	if c.RequestID == "" {
		c.RequestID = generateID()
	}
	if c.Hostname == "" {
		c.Hostname, _ = os.Hostname()
	}
	if c.StartTime.IsZero() {
		c.StartTime = time.Now()
	}
}

type ServerConfig struct {
	Host        string `yaml:"host" default:"localhost"`
	Port        int    `yaml:"port" default:"8080"`
	BindAddress string `yaml:"bind_address"` // Computed from Host:Port
}

// SetDefaults for nested struct
func (s *ServerConfig) SetDefaults() {
	if s.BindAddress == "" {
		s.BindAddress = fmt.Sprintf("%s:%d", s.Host, s.Port)
	}
}

func main() {
	var cfg Config
	if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Configuration with Dynamic Defaults ===")
	fmt.Printf("App Name:   %s\n", cfg.AppName)
	fmt.Printf("Request ID: %s\n", cfg.RequestID)
	fmt.Printf("Hostname:   %s\n", cfg.Hostname)
	fmt.Printf("Start Time: %s\n", cfg.StartTime.Format(time.RFC3339))
	fmt.Printf("Bind Addr:  %s\n", cfg.Server.BindAddress)
}

// generateID creates a simple random ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
