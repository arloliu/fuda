# Fuda (札)

<p align="center">
  <img src="docs/fuda-logo.png" alt="Fuda Logo" width="300">
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/arloliu/fuda.svg)](https://pkg.go.dev/github.com/arloliu/fuda)

> ⛩️ _Spiritual protection and hydration for your Go configurations._

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
- **Dotenv file loading** via `WithDotEnv()` with overlay and override support
- **External references** via `ref` and `refFrom` tags (file://, http://, https://, vault://)
- **DSN composition** via `dsn` tag for building connection strings from fields
- **HashiCorp Vault integration** via `fuda/vault` package (Token, Kubernetes, AppRole auth)
- **Hot-reload configuration** via `fuda/watcher` package with fsnotify
- **Template processing** via Go's `text/template` for dynamic configuration
- **Custom type conversion** via `Scanner` interface
- **Dynamic defaults** via `Setter` interface
- **Duration type** with human-friendly parsing (e.g., `"30s"`, `"5m"`, `"1h30m"`, `"7d"`)
- **RawMessage type** for deferred/polymorphic JSON/YAML unmarshaling
- **Validation** using [go-playground/validator](https://github.com/go-playground/validator)

## Documentation

- **[User Guide](docs/user-guide.md)** - Complete guide with examples for all features
- **[Tag Specification](docs/tag-spec.md)** - Complete reference for all struct tags
- **[Setter & Scanner](docs/setter-scanner.md)** - Custom type conversion and dynamic defaults
- **[Custom Resolvers](docs/custom-resolvers.md)** - Implementing custom reference resolvers
- **[Vault Resolver](vault/README.md)** - HashiCorp Vault integration (separate module: `go get github.com/arloliu/fuda/vault`)
- **[Config Watcher](docs/config-watcher.md)** - Hot-reload configuration watching

## API

### Builder Pattern

```go
loader, err := fuda.New().
    FromFile("config.yaml").           // or FromBytes(), FromReader()
    WithEnvPrefix("APP_").             // optional: prefix for env vars
    WithDotEnv(".env").                // optional: load .env file
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

````yaml
host: "{{ .Host }}"
Custom delimiters can be set if your config contains literal `{{` sequences:

```go
WithTemplate(data, fuda.WithDelimiters("<{", "}>"))
````

### Dotenv Loading

Load environment variables from `.env` files before processing:

```go
// Single file
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithDotEnv(".env").
    Build()

// Multiple files (overlay pattern)
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithDotEnvFiles([]string{".env", ".env.local", ".env.production"}).
    Build()

// Override mode (dotenv values override existing env vars)
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithDotEnv(".env", fuda.DotEnvOverride()).
    Build()
```

Missing files are silently ignored, making this safe for optional overlays like `.env.local`.

### DSN Composition

Build connection strings from config fields and secrets using the `dsn` tag:

```go
type Config struct {
    DBHost     string `yaml:"host" default:"localhost"`
    DBUser     string `env:"DB_USER"`
    DBPassword string `ref:"vault:///secret/data/db#password"`

    // Compose DSN from fields above, or inline secrets/env vars
    DSN string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:5432/app"`
}
```

Inline secret and environment variable resolution:

```go
// Inline vault secret
DSN string `dsn:"postgres://${ref:vault:///db#user}:${ref:vault:///db#pass}@host:5432/db"`

// Inline environment variables
DSN string `dsn:"redis://${env:REDIS_HOST}:${env:REDIS_PORT}/0"`
```

### Convenience Functions

```go
// Load from file
fuda.LoadFile("config.yaml", &cfg)

// Load from bytes
fuda.LoadBytes(yamlData, &cfg)

// Load from reader
fuda.LoadReader(reader, &cfg)
```

## Important Notes

| Topic            | Details                                                                            |
| ---------------- | ---------------------------------------------------------------------------------- |
| **Timeout**      | Default is `0` (no timeout). Set explicitly with `WithTimeout()` for network refs. |
| **Blocking I/O** | `Load()` blocks during file/network operations. Run in a goroutine if needed.      |
| **File URIs**    | Supports `file:///absolute/path` and `file://relative/path` formats.               |

## Thread Safety

| Component     | Thread-Safe?                                                            |
| ------------- | ----------------------------------------------------------------------- |
| `Loader`      | ✅ Yes — safe to call `Load()` from multiple goroutines after `Build()` |
| `RefResolver` | ⚠️ Implementations must be thread-safe if `Loader` is shared            |

A `Loader` instance does not mutate state after construction and can be safely reused.

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

| Error Type         | When Returned                                                         |
| ------------------ | --------------------------------------------------------------------- |
| `*FieldError`      | Invalid tag value, type conversion failure, or ref resolution failure |
| `*LoadError`       | Multiple field errors during a single load operation                  |
| `*ValidationError` | Validation rules from `validate` tag failed                           |

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

## Examples

Working examples are available in the [`examples/`](examples/) directory:

| Example                            | Description                                                 |
| ---------------------------------- | ----------------------------------------------------------- |
| [basic](examples/basic/)           | Loading config with defaults, env overrides, and validation |
| [dotenv](examples/dotenv/)         | Loading environment variables from .env files               |
| [dsn](examples/dsn/)               | Composing connection strings using `dsn` tag                |
| [refs](examples/refs/)             | External references with `ref` and `refFrom` tags           |
| [scanner](examples/scanner/)       | Custom type conversion via `Scanner` interface              |
| [setter](examples/setter/)         | Dynamic defaults via `Setter` interface                     |
| [template](examples/template/)     | Go template processing for dynamic config                   |
| [validation](examples/validation/) | Struct validation with `validate` tag                       |
| [watcher](examples/watcher/)       | Hot-reload configuration with fsnotify                      |
