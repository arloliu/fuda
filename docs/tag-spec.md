# Tag Reference

Quick reference for all fuda struct tags.

## At a Glance

```go
type Config struct {
    Host     string        `yaml:"host" default:"localhost" env:"APP_HOST" validate:"required"`
    Password string        `ref:"file:///run/secrets/db_pass"`
    Token    string        `refFrom:"TokenPath"`
    Timeout  time.Duration `default:"30s"`
    DSN      string        `dsn:"postgres://${.User}:${.Password}@${.Host}:5432/db"`
}
```

| Tag           | Purpose                               | Priority      |
| ------------- | ------------------------------------- | ------------- |
| `env`         | Environment variable override         | Highest       |
| `yaml`/`json` | Config file key                       | -             |
| `ref`         | Load from fixed URI                   | -             |
| `refFrom`     | Load from URI in another field        | -             |
| `default`     | Fallback value                        | Lowest        |
| `dsn`         | Compose connection string from fields | After default |
| `validate`    | Validation rules                      | After loading |

**Priority order:** `env` > config file > `ref`/`refFrom` > `default` > `dsn`

---

## `default` Tag

Sets a fallback value when the field is zero after parsing.

```go
Host    string        `default:"localhost"`
Port    int           `default:"8080"`
Enabled *bool         `default:"true"`
Timeout time.Duration `default:"30s"`
Tags    []string      `default:"[\"app\", \"prod\"]"`
```

### Supported Types

| Type                               | Example                                         |
| ---------------------------------- | ----------------------------------------------- |
| `string`, `int`, `bool`, `float64` | `default:"value"`                               |
| `*string`, `*int`, etc.            | `default:"value"`                               |
| `time.Duration`                    | `default:"30s"`, `default:"5m"`, `default:"7d"` |
| `time.Time`                        | `default:"2024-01-01T00:00:00Z"` (RFC3339)      |
| `[]T`                              | `default:"[1, 2, 3]"` (JSON array)              |
| `map[K]V`                          | `default:"{\"key\": \"value\"}"` (JSON object)  |

### Duration Parsing

Fuda extends Go's standard `time.ParseDuration` to support **days** with the `d` suffix:

| Unit         | Suffix    | Example              |
| ------------ | --------- | -------------------- |
| Days         | `d`       | `7d` (7 days = 168h) |
| Hours        | `h`       | `24h`                |
| Minutes      | `m`       | `30m`                |
| Seconds      | `s`       | `45s`                |
| Milliseconds | `ms`      | `500ms`              |
| Microseconds | `us`/`µs` | `100us`              |
| Nanoseconds  | `ns`      | `1000ns`             |

Units can be combined: `1d12h30m` (1 day, 12 hours, 30 minutes). Fractional days are supported: `0.5d` (12 hours).

### Skip Default

Use `-` to skip default processing:

```go
Field string `default:"-"`
```

---

## `env` Tag

Maps a field to an environment variable.

```go
Host string `env:"DB_HOST"`
Port int    `env:"DB_PORT"`
```

### With Prefix

```go
loader, _ := fuda.New().
    FromFile("config.yaml").
    WithEnvPrefix("APP_").
    Build()
```

With prefix `APP_`:

- `env:"HOST"` reads from `APP_HOST`
- `env:"PORT"` reads from `APP_PORT`

---

## `ref` Tag

Loads a value from a fixed URI (only if field is zero).

```go
// From file
Password string `ref:"file:///run/secrets/db_password"`

// From HTTP
APIKey string `ref:"https://vault.example.com/v1/secrets/api_key"`
```

### Supported Schemes

| Scheme     | Description    |
| ---------- | -------------- |
| `file://`  | Local file     |
| `http://`  | HTTP endpoint  |
| `https://` | HTTPS endpoint |

---

## `refFrom` Tag

Loads a value from a URI stored in another field.

```go
type Config struct {
    TokenPath string `yaml:"token_path"`
    Token     string `refFrom:"TokenPath"`
}
```

```yaml
token_path: '/run/secrets/app_token'
```

### Path Normalization

Bare paths are automatically prefixed with `file://`:

| Input                 | Normalized                  |
| --------------------- | --------------------------- |
| `/run/secrets/token`  | `file:///run/secrets/token` |
| `file:///path`        | `file:///path`              |
| `https://example.com` | `https://example.com`       |

---

## `dsn` Tag

Composes connection strings from other fields using template syntax with `${...}` delimiters.

The `dsn` tag is processed **after** all other tags (`env`, `ref`, `default`), so referenced fields have their final values.

### Template Syntax

| Syntax             | Description                         |
| ------------------ | ----------------------------------- |
| `${.FieldName}`    | Value of a field in the same struct |
| `${.Nested.Field}` | Nested struct field access          |
| `${ref:uri}`       | Resolve a URI inline                |
| `${env:KEY}`       | Read an environment variable        |

---

### Example: Using Field References

Compose DSN from other config fields:

```go
type Config struct {
    DBHost     string `yaml:"host" default:"localhost"`
    DBPort     int    `yaml:"port" default:"5432"`
    DBName     string `yaml:"name" default:"myapp"`
    DBUser     string `yaml:"user"`
    DBPassword string `yaml:"password"`

    // Compose from fields above
    PostgresDSN string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:${.DBPort}/${.DBName}"`
}
```

---

### Example: Using `${ref:uri}` for Secrets

Resolve secrets inline without storing them in fields:

```go
type Config struct {
    DBHost string `yaml:"host" default:"localhost"`

    // Inline vault secret resolution
    DSN string `dsn:"postgres://${ref:vault:///secret/data/db#user}:${ref:vault:///secret/data/db#pass}@${.DBHost}:5432/app"`
}
```

Supported URI schemes for `${ref:...}`:

- `file://` - Local files (e.g., Docker secrets)
- `http://` / `https://` - HTTP endpoints
- `vault://` - HashiCorp Vault (requires vault resolver)

```go
// Docker secrets
DSN string `dsn:"postgres://admin:${ref:file:///run/secrets/db_password}@db:5432/app"`

// File-based secret
DSN string `dsn:"redis://:${ref:file://./secrets/redis.txt}@redis:6379/0"`
```

---

### Example: Using `${env:KEY}` for Environment Variables

Read environment variables inline:

```go
type Config struct {
    DBHost string `yaml:"host" default:"localhost"`

    // Mix env vars with field references
    DSN string `dsn:"postgres://${env:DB_USER}:${env:DB_PASSWORD}@${.DBHost}:5432/app"`
}
```

Environment variables respect the `WithEnvPrefix()` setting:

```go
// With WithEnvPrefix("APP_"):
DSN string `dsn:"postgres://${env:DB_USER}@localhost:5432/db"`
// ${env:DB_USER} reads APP_DB_USER
```

---

### Example: Nested Struct Fields

Access fields from nested structs:

```go
type Config struct {
    Database struct {
        Host string `yaml:"host" default:"localhost"`
        User string `yaml:"user"`
        Pass string `yaml:"password"`
    } `yaml:"database"`

    Redis struct {
        Host string `yaml:"host" default:"localhost"`
        Port int    `yaml:"port" default:"6379"`
    } `yaml:"redis"`

    PostgresDSN string `dsn:"postgres://${.Database.User}:${.Database.Pass}@${.Database.Host}:5432/app"`
    RedisDSN    string `dsn:"redis://${.Redis.Host}:${.Redis.Port}/0"`
}
```

---

### Example: Mixed Sources

Combine field references, secrets, and env vars:

```go
type Config struct {
    DBHost string `yaml:"host" env:"DB_HOST" default:"localhost"`
    DBPort int    `yaml:"port" default:"5432"`
    DBName string `yaml:"name" default:"production"`

    // User from env, password from vault, host/port from config
    DSN string `dsn:"postgres://${env:DB_USER}:${ref:vault:///secret/db#password}@${.DBHost}:${.DBPort}/${.DBName}?sslmode=require"`
}
```

---

### Strict Mode

By default, empty values produce empty strings (permissive). Enable strict mode to error on empty values:

```go
DSN string `dsn:"postgres://${.User}@${.Host}:5432/db" dsnStrict:"true"`
```

---

## `validate` Tag

Validation rules using [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator/v10).

```go
Host string `validate:"required"`
Port int    `validate:"required,min=1,max=65535"`
Env  string `validate:"required,oneof=dev staging prod"`
```

### Common Rules

| Rule          | Description                  |
| ------------- | ---------------------------- |
| `required`    | Must not be zero             |
| `min=N`       | Minimum value/length         |
| `max=N`       | Maximum value/length         |
| `oneof=a b c` | Must be one of listed values |
| `url`         | Must be valid URL            |
| `email`       | Must be valid email          |

---

## `Setter` Interface

For dynamic defaults that can't be expressed as static strings:

```go
type Config struct {
    RequestID string
}

func (c *Config) SetDefaults() {
    if c.RequestID == "" {
        c.RequestID = uuid.New().String()
    }
}
```

`SetDefaults()` is called after all tag processing.

---

## Nested Structs

All tags work with nested structs:

```go
type Config struct {
    Database DatabaseConfig `yaml:"database"`
}

type DatabaseConfig struct {
    Host string `default:"localhost" env:"DB_HOST"`
    Port int    `default:"5432" env:"DB_PORT"`
}
```

### Optional Sections

Use pointers for optional nested structs:

```go
type Config struct {
    Redis *RedisConfig `yaml:"redis,omitempty"`
}
```

- If `redis` is missing in config → `Redis` is `nil`
- If `redis` is present → tags are processed
