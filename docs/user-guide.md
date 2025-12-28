# Fuda User Guide

This guide covers fuda's features with practical examples. For quick reference, see [Tag Specification](tag-spec.md).

---

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Tag System Deep Dive](#tag-system-deep-dive)
4. [External References](#external-references)
5. [DSN Composition](#dsn-composition)
6. [Template Processing](#template-processing)
7. [Dotenv Support](#dotenv-support)
8. [Validation](#validation)
9. [Custom Type Conversion (Scanner)](#custom-type-conversion-scanner)
10. [Dynamic Defaults (Setter)](#dynamic-defaults-setter)
11. [Vault Integration](#vault-integration)
12. [Hot-Reload Configuration](#hot-reload-configuration)
13. [Error Handling](#error-handling)
14. [Real-World Patterns](#real-world-patterns)
15. [FAQ / Troubleshooting](#faq--troubleshooting)

---

## Introduction

Fuda is a struct-tag-first configuration library for Go. Define your configuration as a struct with tags, and fuda handles:

- **Defaults** — Static fallback values via `default` tag
- **Environment overrides** — Via `env` tag with optional prefix
- **External secrets** — Via `ref`/`refFrom` tags (file, HTTP, Vault)
- **Connection strings** — Via `dsn` tag composition
- **Validation** — Via go-playground/validator integration

### Core Concept: Processing Order

Tags are processed in a specific priority order:

```
env → config file → ref/refFrom → default → dsn → SetDefaults() → validate
```

| Priority    | Source          | When Used                   |
| ----------- | --------------- | --------------------------- |
| 1 (Highest) | `env` tag       | Environment variable is set |
| 2           | Config file     | Field present in YAML/JSON  |
| 3           | `ref`/`refFrom` | Field is still zero         |
| 4           | `default` tag   | Field is still zero         |
| 5           | `dsn` tag       | After all above complete    |
| 6           | `SetDefaults()` | After tags processed        |
| 7 (Lowest)  | `validate` tag  | Final validation            |

---

## Getting Started

### Installation

```bash
go get github.com/arloliu/fuda
```

### First Config Struct

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
    Debug   bool          `yaml:"debug" default:"false" env:"APP_DEBUG"`
}

func main() {
    var cfg Config
    if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
}
```

### Loading Methods

```go
// From file (auto-detects YAML/JSON by extension)
fuda.LoadFile("config.yaml", &cfg)

// From bytes
fuda.LoadBytes(yamlData, &cfg)

// From reader
fuda.LoadReader(reader, &cfg)

// Builder pattern for advanced options
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithEnvPrefix("APP_").
    Build()
loader.Load(&cfg)
```

---

## Tag System Deep Dive

### `default` Tag

Sets a fallback value when the field is zero after other sources are checked.

```go
type Config struct {
    // Primitives
    Host    string  `default:"localhost"`
    Port    int     `default:"8080"`
    Rate    float64 `default:"1.5"`
    Enabled bool    `default:"true"`

    // Duration (supports day units: 7d, 1d12h)
    Timeout time.Duration `default:"30s"`
    MaxAge  time.Duration `default:"7d"`

    // Pointers (creates non-nil value)
    MaxConn *int `default:"100"`

    // Slices and Maps (JSON format)
    Tags   []string          `default:"[\"app\", \"prod\"]"`
    Labels map[string]string `default:"{\"env\": \"prod\"}"`
}
```

**Skip default processing:**

```go
Field string `default:"-"`  // Never apply default
```

### `env` Tag

Maps a field to an environment variable. Environment values have the **highest priority**.

```go
type Config struct {
    Host string `env:"DB_HOST"`  // Reads $DB_HOST
    Port int    `env:"DB_PORT"`  // Reads $DB_PORT, converts to int
}
```

**With prefix:**

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithEnvPrefix("MYAPP_").  // env:"HOST" reads $MYAPP_HOST
    Build()
```

### Processing Priority Example

Consider this config:

```go
type Config struct {
    Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
}
```

| Condition                                             | Result            |
| ----------------------------------------------------- | ----------------- |
| `APP_HOST=api.example.com`, YAML has `host: db.local` | `api.example.com` |
| `APP_HOST` unset, YAML has `host: db.local`           | `db.local`        |
| `APP_HOST` unset, no YAML field                       | `localhost`       |

→ See [Tag Specification](tag-spec.md) for complete reference.

---

## External References

Use `ref` and `refFrom` tags to load values from external sources.

### `ref` Tag — Static URI

```go
type Config struct {
    // Docker secrets
    Password string `ref:"file:///run/secrets/db_password"`

    // HTTP endpoint
    APIKey string `ref:"https://config.example.com/api-key"`

    // Local file
    License string `ref:"file://./license.txt"`
}
```

### Dynamic URI with Templates

Compose URIs from other fields using `${.FieldName}` syntax:

```go
type Config struct {
    SecretsDir string `yaml:"secrets_dir" default:"/etc/secrets"`
    Env        string `yaml:"env" default:"dev"`

    // Dynamic path: /etc/secrets/dev/db_password
    Password string `ref:"file://${.SecretsDir}/${.Env}/db_password"`
}
```

> **Important:** Referenced fields must be declared **earlier** in the struct.

### `refFrom` Tag — Dynamic Path

When the URI itself comes from configuration:

```go
type Config struct {
    TokenPath string `yaml:"token_path"`  // e.g., "/run/secrets/token"
    Token     string `refFrom:"TokenPath"`
}
```

Bare paths are auto-prefixed with `file://`:
| Input | Normalized |
|-------|------------|
| `/run/secrets/token` | `file:///run/secrets/token` |
| `https://example.com/key` | `https://example.com/key` |

### Timeout for Network Requests

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithTimeout(10 * time.Second).
    Build()
```

→ See [refs example](../examples/refs/) for runnable code.

---

## DSN Composition

Build connection strings from config fields using the `dsn` tag. DSN is processed **after** all other tags, so fields have their final values.

### Basic Example

```go
type Config struct {
    DBHost     string `yaml:"host" default:"localhost"`
    DBPort     int    `yaml:"port" default:"5432"`
    DBUser     string `env:"DB_USER"`
    DBPassword string `env:"DB_PASSWORD"`

    // Compose from fields above
    DSN string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:${.DBPort}/mydb"`
}
```

### Inline Secrets and Env Vars

```go
type Config struct {
    Host string `default:"localhost"`

    // Inline environment variable
    RedisDSN string `dsn:"redis://:${env:REDIS_PASS}@${.Host}:6379/0"`

    // Inline file reference (Docker secret)
    MongoDSN string `dsn:"mongodb://admin:${ref:file:///run/secrets/mongo_pass}@${.Host}:27017/db"`

    // Inline Vault secret
    PostgresDSN string `dsn:"postgres://${ref:vault:///secret/db#user}:${ref:vault:///secret/db#pass}@${.Host}/db"`
}
```

### Strict Mode

Error on empty values instead of silently producing empty strings:

```go
DSN string `dsn:"postgres://${.User}@${.Host}/db" dsnStrict:"true"`
```

→ See [dsn example](../examples/dsn/) for runnable code.

---

## Template Processing

Process configuration files as Go templates before YAML/JSON parsing.

### Basic Usage

```go
type TemplateData struct {
    Env     string
    Version string
}

data := TemplateData{Env: "prod", Version: "1.2.3"}

loader, _ := fuda.New().
    FromFile("config.yaml").
    WithTemplate(data).
    Build()
```

**config.yaml:**

```yaml
environment: "{{ .Env }}"
version: "{{ .Version }}"
feature_flags:
  {{ if eq .Env "prod" }}
  debug: false
  {{ else }}
  debug: true
  {{ end }}
```

### Custom Delimiters

If your config contains literal `{{`:

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithTemplate(data, fuda.WithDelimiters("<{", "}>")).
    Build()
```

→ See [template example](../examples/template/) for runnable code.

---

## Dotenv Support

Load `.env` files before configuration processing.

### Single File

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithDotEnv(".env").
    Build()
```

### Overlay Pattern

Load multiple files, later files override earlier ones:

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithDotEnvFiles([]string{
        ".env",            // Base config
        ".env.local",      // Local overrides (gitignored)
        ".env.production", // Environment-specific
    }).
    Build()
```

Missing files are silently ignored — perfect for optional `.env.local` files.

### Override Mode

By default, `.env` values only set vars if not already defined. Use override mode to always apply:

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithDotEnv(".env.production", fuda.DotEnvOverride()).
    Build()
```

→ See [dotenv example](../examples/dotenv/) for runnable code.

---

## Validation

Validate configuration using [go-playground/validator](https://github.com/go-playground/validator) rules.

```go
type Config struct {
    Host     string `validate:"required,hostname"`
    Port     int    `validate:"required,min=1,max=65535"`
    LogLevel string `validate:"required,oneof=debug info warn error"`
    Email    string `validate:"omitempty,email"`  // optional, but must be valid if present
    Timeout  int    `validate:"gte=0,lte=300"`
}
```

### Common Validation Rules

| Rule             | Description                                  |
| ---------------- | -------------------------------------------- |
| `required`       | Must not be zero value                       |
| `min=N`, `max=N` | Minimum/maximum for numbers or string length |
| `oneof=a b c`    | Must be one of the listed values             |
| `url`, `email`   | Format validation                            |
| `gte=N`, `lte=N` | Greater/less than or equal                   |

### Custom Validator

```go
import "github.com/go-playground/validator/v10"

v := validator.New()
v.RegisterValidation("myRule", myValidationFunc)

loader, _ := fuda.New().
    FromFile("config.yaml").
    WithValidator(v).
    Build()
```

→ See [validation example](../examples/validation/) for runnable code.

---

## Custom Type Conversion (Scanner)

Implement the `Scanner` interface for custom string-to-value conversion from `default` tags.

```go
type Scanner interface {
    Scan(src any) error
}
```

### Example: Log Level Enum

```go
type LogLevel int

const (
    Debug LogLevel = iota
    Info
    Warn
    Error
)

func (l *LogLevel) Scan(src any) error {
    s, ok := src.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", src)
    }
    switch strings.ToLower(s) {
    case "debug": *l = Debug
    case "info":  *l = Info
    case "warn":  *l = Warn
    case "error": *l = Error
    default:
        return fmt.Errorf("unknown log level: %s", s)
    }
    return nil
}

// Usage
type Config struct {
    Level LogLevel `default:"info"`
}
```

→ See [Setter & Scanner Guide](setter-scanner.md) for more examples.

---

## Dynamic Defaults (Setter)

Implement the `Setter` interface for defaults that require computation.

```go
type Setter interface {
    SetDefaults()
}
```

`SetDefaults()` is called **after** all tag processing but **before** validation.

### Example: Computed Values

```go
type Config struct {
    Host      string `default:"localhost"`
    Port      int    `default:"8080"`
    BaseURL   string // Computed
    RequestID string // Dynamic
}

func (c *Config) SetDefaults() {
    if c.BaseURL == "" {
        c.BaseURL = fmt.Sprintf("http://%s:%d", c.Host, c.Port)
    }
    if c.RequestID == "" {
        c.RequestID = uuid.New().String()
    }
}
```

### Nested Structs

`SetDefaults` is called in post-order (children before parents):

```go
type App struct {
    DB     Database
    Server Server
}

func (d *Database) SetDefaults() { /* Called first */ }
func (s *Server) SetDefaults()   { /* Called second */ }
func (a *App) SetDefaults()      { /* Called last — can use child values */ }
```

→ See [Setter & Scanner Guide](setter-scanner.md) for details.

---

## Vault Integration

The `fuda/vault` package provides HashiCorp Vault integration as a separate module.

### Installation

```bash
go get github.com/arloliu/fuda/vault
```

### Quick Start

```go
import (
    "github.com/arloliu/fuda"
    "github.com/arloliu/fuda/vault"
)

// Create resolver
resolver, _ := vault.NewResolver(
    vault.WithAddress("https://vault.example.com:8200"),
    vault.WithToken(os.Getenv("VAULT_TOKEN")),
)

// Use with fuda
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithRefResolver(resolver).
    Build()
```

### URI Format

```
vault:///<mount>/data/<path>#<field>
```

```go
type Config struct {
    DBPassword string `ref:"vault:///secret/data/myapp#db_password"`
    APIKey     string `ref:"vault:///secret/data/myapp#api_key"`
}
```

### Authentication Methods

```go
// Token (dev/CI)
vault.WithToken(os.Getenv("VAULT_TOKEN"))

// Kubernetes (pods)
vault.WithKubernetesAuth("my-role", "/var/run/secrets/.../token")

// AppRole (machine-to-machine)
vault.WithAppRole(roleID, secretID)
```

→ See [Vault README](../vault/README.md) for complete documentation.

---

## Hot-Reload Configuration

The `fuda/watcher` package enables automatic configuration reloading.

### Quick Start

```go
import (
    "sync/atomic"
    "github.com/arloliu/fuda/watcher"
)

var globalConfig atomic.Pointer[Config]

func main() {
    w, _ := watcher.New().
        FromFile("config.yaml").
        WithWatchInterval(30 * time.Second).
        Build()
    defer w.Stop()

    var cfg Config
    updates, _ := w.Watch(&cfg)
    globalConfig.Store(&cfg)

    // Handle updates
    go func() {
        for newCfg := range updates {
            globalConfig.Store(newCfg.(*Config))
        }
    }()

    // Use config safely
    config := globalConfig.Load()
    fmt.Println(config.Host)
}
```

### Watch Mechanisms

| Source                      | Mechanism                       |
| --------------------------- | ------------------------------- |
| Config files, local secrets | fsnotify (real-time)            |
| Vault, HTTP refs            | Polling (configurable interval) |

→ See [Config Watcher Guide](config-watcher.md) for details.

---

## Error Handling

Fuda provides typed errors for inspection.

### Error Types

| Type               | When Returned                                        |
| ------------------ | ---------------------------------------------------- |
| `*FieldError`      | Tag parsing, type conversion, ref resolution failure |
| `*LoadError`       | Multiple field errors in one load                    |
| `*ValidationError` | Validation rules failed                              |

### Inspecting Errors

```go
var fieldErr *fuda.FieldError
if errors.As(err, &fieldErr) {
    fmt.Printf("Field: %s, Tag: %s, Cause: %v\n",
        fieldErr.Path, fieldErr.Tag, fieldErr.Unwrap())
}

var loadErr *fuda.LoadError
if errors.As(err, &loadErr) {
    for _, e := range loadErr.Unwrap() {
        // Handle each field error
    }
}

var validationErr *fuda.ValidationError
if errors.As(err, &validationErr) {
    // Access validation details
}
```

---

## Real-World Patterns

### Web Server Configuration

```go
type Config struct {
    // Server
    Host         string        `yaml:"host" default:"0.0.0.0" env:"HOST"`
    Port         int           `yaml:"port" default:"8080" env:"PORT"`
    ReadTimeout  time.Duration `yaml:"read_timeout" default:"30s"`
    WriteTimeout time.Duration `yaml:"write_timeout" default:"30s"`

    // Database
    Database struct {
        Host     string `yaml:"host" default:"localhost" env:"DB_HOST"`
        Port     int    `yaml:"port" default:"5432" env:"DB_PORT"`
        Name     string `yaml:"name" default:"app" env:"DB_NAME"`
        User     string `env:"DB_USER" validate:"required"`
        Password string `ref:"file:///run/secrets/db_password"`
    } `yaml:"database"`

    // Computed DSN
    DatabaseDSN string `dsn:"postgres://${.Database.User}:${.Database.Password}@${.Database.Host}:${.Database.Port}/${.Database.Name}?sslmode=require"`

    // Logging
    LogLevel string `yaml:"log_level" default:"info" env:"LOG_LEVEL" validate:"oneof=debug info warn error"`
}

func (c *Config) SetDefaults() {
    // Add request tracing ID
    if c.LogLevel == "debug" {
        log.Println("Debug mode enabled")
    }
}
```

### Multi-Environment with Dotenv

```bash
# .env (committed, base values)
DB_HOST=localhost
DB_PORT=5432
LOG_LEVEL=info

# .env.local (gitignored, developer overrides)
DB_HOST=docker.local
LOG_LEVEL=debug

# .env.production (deployed)
DB_HOST=prod-db.internal
LOG_LEVEL=warn
```

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithDotEnvFiles([]string{".env", ".env.local"}).
    Build()
```

### Microservice with Vault and Hot-Reload

```go
func main() {
    // Setup Vault resolver
    resolver, _ := vault.NewResolver(
        vault.WithAddress(os.Getenv("VAULT_ADDR")),
        vault.WithKubernetesAuth("my-service", serviceAccountTokenPath),
    )

    // Create watcher with Vault support
    w, _ := watcher.New().
        FromFile("/etc/config/app.yaml").
        WithRefResolver(resolver).
        WithWatchInterval(5 * time.Minute).
        Build()
    defer w.Stop()

    var cfg Config
    updates, _ := w.Watch(&cfg)
    globalConfig.Store(&cfg)

    go func() {
        for newCfg := range updates {
            log.Println("Config reloaded")
            globalConfig.Store(newCfg.(*Config))
        }
    }()

    // Start server with config...
}
```

---

## FAQ / Troubleshooting

### Q: My environment variable isn't being read

**Check:**

1. Variable is set before application starts
2. `env` tag matches exactly (case-sensitive)
3. If using prefix, variable includes prefix: `env:"HOST"` with `WithEnvPrefix("APP_")` reads `APP_HOST`

### Q: My `ref` tag returns empty

**Check:**

1. File path is correct and readable
2. For absolute paths: use `file:///absolute/path`
3. For relative paths: use `file://./relative/path`
4. Set `WithTimeout()` for network requests

### Q: Field order matters for template references

Template expressions like `${.FieldName}` can only reference fields **declared earlier** in the struct. Reorder your fields so dependencies come first.

### Q: Validation fails but value looks correct

The `validate` tag runs **after** all other tags. Check if:

1. Value is actually set (not empty string or zero)
2. Type conversion succeeded
3. `required` is used with `omitempty` for optional fields

### Best Practices Checklist

- [ ] Use `env` for deployment-specific values (ports, hosts)
- [ ] Use `default` for sensible development values
- [ ] Use `ref` for secrets (never hardcode in config files)
- [ ] Use `validate:"required"` for mandatory values
- [ ] Use pointer types for optional nested structs
- [ ] Set explicit timeouts for network refs
- [ ] Use `atomic.Pointer` for thread-safe hot-reload access
