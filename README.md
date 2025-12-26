# Fuda (札)

<p align="center">
  <img src="docs/fuda-logo.png" alt="Fuda Logo" width="300">
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/arloliu/fuda.svg)](https://pkg.go.dev/github.com/arloliu/fuda)

> ⛩️ *Spiritual protection and hydration for your Go configurations.*

**Fuda** is a lightweight, struct-tag-first configuration library for Go — with built-in defaults, environment overrides, secret references, and validation.

## Why Fuda?

The name comes from **御札 (ofuda)** — traditional Japanese talismans inscribed with sacred characters to ward off evil and bring protection. Just as an ofuda guards a home, fuda guards your application:

- 📜 **Inscribe your struct fields** with tags like `default`, `env`, and `ref`
- 🔮 **Summon configuration** from files, environment, or remote URLs
- 🛡️ **Protect integrity** with validation rules
- 🔐 **Guard secrets** by resolving them securely at runtime (Docker secrets, vaults)

No more scattered `os.Getenv()` calls. No more manual YAML wrangling. Just declare your config struct, add the sacred tags, and let fuda perform the ritual.

## Installation

```bash
go get github.com/arloliu/fuda
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/arloliu/fuda"
)

type Config struct {
    Host    string        `yaml:"host" default:"localhost" env:"APP_HOST"`
    Port    int           `yaml:"port" default:"8080" env:"APP_PORT"`
    Timeout time.Duration `yaml:"timeout" default:"30s"`
    Debug   bool          `yaml:"debug" default:"false"`
}

func main() {
    var cfg Config

    // Load from file
    if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
}
```

## Features

- **YAML/JSON parsing** with struct tag support
- **Default values** via `default` tag
- **Environment overrides** via `env` tag with optional prefix
- **External references** via `ref` and `refFrom` tags (file://, http://, https://, vault://)
- **HashiCorp Vault integration** via `fuda/vault` package (Token, Kubernetes, AppRole auth)
- **Hot-reload configuration** via `fuda/watcher` package with fsnotify
- **Template processing** via Go's `text/template` for dynamic configuration
- **Custom type conversion** via `Scanner` interface
- **Dynamic defaults** via `Setter` interface
- **Validation** using [go-playground/validator](https://github.com/go-playground/validator)

## API

### Builder Pattern

```go
loader, err := fuda.New().
    FromFile("config.yaml").           // or FromBytes(), FromReader()
    WithEnvPrefix("APP_").             // optional: prefix for env vars
    WithTimeout(10 * time.Second).     // optional: timeout for ref resolution
    WithValidator(customValidator).    // optional: custom validator
    WithRefResolver(customResolver).   // optional: custom ref resolver
    WithTemplate(templateData).        // optional: template processing
    Build()

if err != nil {
    log.Fatal(err)
}

var cfg Config
if err := loader.Load(&cfg); err != nil {
    log.Fatal(err)
}
```

### Template Processing

Process configuration as a Go template before parsing:

```go
type TemplateData struct {
    Env  string
    Host string
}

data := TemplateData{Env: "prod", Host: "api.example.com"}

loader, _ := fuda.New().
    FromFile("config.yaml").
    WithTemplate(data).
    Build()
```

With `config.yaml`:
```yaml
host: "{{ .Host }}"
environment: "{{ .Env }}"
```

Custom delimiters can be set if your config contains literal `{{` sequences:

```go
WithTemplate(data, fuda.WithDelimiters("<{", "}>"))
```

### Convenience Functions
g
```go
// Load from file
fuda.LoadFile("config.yaml", &cfg)

// Load from bytes
fuda.LoadBytes(yamlData, &cfg)

// Load from reader
fuda.LoadReader(reader, &cfg)
```

## Important Notes

| Topic | Details |
|-------|---------|
| **Timeout** | Default is `0` (no timeout). Set explicitly with `WithTimeout()` for network refs. |
| **Blocking I/O** | `Load()` blocks during file/network operations. Run in a goroutine if needed. |
| **File URIs** | Supports `file:///absolute/path` and `file://relative/path` formats. |

## Thread Safety

| Component | Thread-Safe? |
|-----------|--------------|
| `Loader` | ✅ Yes — safe to call `Load()` from multiple goroutines after `Build()` |
| `RefResolver` | ⚠️ Implementations must be thread-safe if `Loader` is shared |

A `Loader` instance does not mutate state after construction and can be safely reused.

## Documentation

- **[Tag Specification](docs/tag-spec.md)** - Complete reference for all struct tags
- **[Setter & Scanner](docs/setter-scanner.md)** - Custom type conversion and dynamic defaults
- **[Custom Resolvers](docs/custom-resolvers.md)** - Implementing custom reference resolvers
- **[Vault Resolver](vault/README.md)** - HashiCorp Vault integration (separate module: `go get github.com/arloliu/fuda/vault`)
- **[Config Watcher](docs/config-watcher.md)** - Hot-reload configuration watching

### Setter Interface

Implement `Setter` for dynamic defaults that can't be expressed as static tag values:

```go
type Config struct {
    RequestID string
    CreatedAt time.Time
}

func (c *Config) SetDefaults() {
    if c.RequestID == "" {
        c.RequestID = uuid.New().String()
    }
    if c.CreatedAt.IsZero() {
        c.CreatedAt = time.Now()
    }
}
```

### Scanner Interface

Implement `Scanner` for custom string-to-value conversion in `default` tags:

```go
type LogLevel int

const (Debug LogLevel = iota; Info; Warn; Error)

func (l *LogLevel) Scan(src any) error {
    s, _ := src.(string)
    switch strings.ToLower(s) {
    case "debug": *l = Debug
    case "info":  *l = Info
    case "warn":  *l = Warn
    case "error": *l = Error
    default: return fmt.Errorf("unknown level: %s", s)
    }
    return nil
}

// Usage
type Config struct {
    Level LogLevel `default:"info"`
}
```

## Error Handling

The library provides typed errors for inspection:

| Error Type | When Returned |
|------------|---------------|
| `*FieldError` | Invalid tag value, type conversion failure, or ref resolution failure |
| `*LoadError` | Multiple field errors during a single load operation |
| `*ValidationError` | Validation rules from `validate` tag failed |

All errors support `errors.Is()` and `errors.Unwrap()` for error chain inspection.

```go
var fieldErr *fuda.FieldError
if errors.As(err, &fieldErr) {
    fmt.Printf("Field: %s, Tag: %s\n", fieldErr.Path, fieldErr.Tag)
}

var validationErr *fuda.ValidationError
if errors.As(err, &validationErr) {
    // Handle validation failures
}
```

