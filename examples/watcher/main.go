// Example: Hot-reload configuration with watcher
//
// This example demonstrates using the watcher package
// for automatic configuration reloading.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/arloliu/fuda/watcher"
)

type Config struct {
	AppName  string `yaml:"app_name" default:"watcher-example"`
	LogLevel string `yaml:"log_level" default:"info"`
	MaxConns int    `yaml:"max_connections" default:"100"`
	Timeout  string `yaml:"timeout" default:"30s"`
}

var globalConfig atomic.Pointer[Config]

func main() {
	// Create config file
	createConfigFile()

	w, err := watcher.New().
		FromFile("config.yaml").
		WithWatchInterval(5 * time.Second).           // Poll interval for remote refs
		WithDebounceInterval(100 * time.Millisecond). // Debounce rapid changes
		Build()
	if err != nil {
		log.Fatal(err)
	}
	defer w.Stop()

	var cfg Config
	updates, err := w.Watch(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Store initial config
	globalConfig.Store(&cfg)
	printConfig("Initial config", &cfg)

	// Handle updates in a goroutine
	go func() {
		for newCfg := range updates {
			c := newCfg.(*Config)
			globalConfig.Store(c)
			printConfig("Config updated", c)
		}
	}()

	fmt.Println("\n=== Watching for changes ===")
	fmt.Println("Edit config.yaml to see hot-reload in action!")
	fmt.Println("Press Ctrl+C to exit")

	// Simulate config changes for demo
	go simulateChanges()

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
}

func createConfigFile() {
	content := `app_name: "watcher-demo"
log_level: "info"
max_connections: 100
timeout: "30s"
`
	os.WriteFile("config.yaml", []byte(content), 0o644)
}

func simulateChanges() {
	time.Sleep(3 * time.Second)

	// Simulate a config change
	content := `app_name: "watcher-demo"
log_level: "debug"
max_connections: 200
timeout: "60s"
`
	fmt.Println("[Simulating config update...]")
	os.WriteFile("config.yaml", []byte(content), 0o644)
}

func printConfig(label string, cfg *Config) {
	fmt.Printf("\n%s:\n", label)
	fmt.Printf("  App:      %s\n", cfg.AppName)
	fmt.Printf("  LogLevel: %s\n", cfg.LogLevel)
	fmt.Printf("  MaxConns: %d\n", cfg.MaxConns)
	fmt.Printf("  Timeout:  %s\n", cfg.Timeout)
}
