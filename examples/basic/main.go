// Example: configuration precedence.
package main

import (
	"fmt"
	"log"

	"github.com/arloliu/fuda"
)

// Config reads values from YAML, falls back to defaults, and accepts env overrides.
type Config struct {
	Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
	Port int    `yaml:"port" default:"8080" env:"APP_PORT"`
}

func main() {
	loader, err := fuda.New().FromFile("config.yaml").Build()
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
}
