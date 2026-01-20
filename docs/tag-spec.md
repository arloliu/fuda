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
| `ref`         | Load from URI (supports templates)    | -             |
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
| `[]byte`                           | `default:"raw content"` (raw bytes)             |
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

### Byte Size Parsing

Integer fields (`int`, `int64`, `uint64`, etc.) support human-readable byte size strings with IEC (binary) and SI (decimal) units:

```go
MaxFileSize   int64  `yaml:"max_file_size" default:"10MiB"`
BufferSize    int    `yaml:"buffer_size" default:"64KiB"`
UploadLimit   uint64 `yaml:"upload_limit" default:"2GB"`
```

```yaml
max_file_size: 10MiB    # 10485760 bytes
buffer_size: 64KiB      # 65536 bytes
upload_limit: 2GB       # 2000000000 bytes
```

**Supported Units:**

| Type   | Units                             | Base   |
| ------ | --------------------------------- | ------ |
| IEC    | `B`, `KiB`, `MiB`, `GiB`, `TiB`, `PiB`, `EiB` | 1024   |
| SI     | `B`, `KB`, `MB`, `GB`, `TB`, `PB`, `EB`       | 1000   |

- Units are **case-insensitive** (`kib`, `KiB`, `KIB` all work)
- Decimal values are supported when they resolve to whole bytes: `0.5MiB` = 524288 bytes
- Fractional bytes are rejected: `0.5KiB` works (512 bytes), but `0.1B` fails
- String fields are **not** coerced—use integer types for byte sizes

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

Loads a value from a URI (only if field is zero). Supports [template syntax](#template-syntax) for dynamic URIs.

### Basic Usage

```go
// From file
Password string `ref:"file:///run/secrets/db_password"`

// From HTTP
APIKey string `ref:"https://vault.example.com/v1/secrets/api_key"`

// Binary content (e.g., certificates, keys)
Cert []byte `ref:"file:///etc/ssl/certs/server.pem"`
Key  []byte `ref:"file:///etc/ssl/private/server.key"`
```

### Dynamic URI with Templates

Compose URIs from other fields using `${...}` syntax:

```go
type Config struct {
    SecretDir string `yaml:"secretDir" default:"/etc/secrets"`
    Fab       string `yaml:"fab" validate:"required"`
    Account   string `yaml:"account" validate:"required"`

    // Dynamic path composed from fields above
    Password string `ref:"file://${.SecretDir}/tcs-${.Fab}-${.Account}-password"`
}
```

> **Note:** Referenced fields must appear **earlier** in the struct to have their values available.

### Supported Schemes

| Scheme     | Description          |
| ---------- | -------------------- |
| `file://`  | Local file           |
| `http://`  | HTTP endpoint        |
| `https://` | HTTPS endpoint       |
| `env://`   | Environment variable |

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

> **Note:** The referenced field must be a **string** type.

### Path Normalization

Bare paths are automatically prefixed with `file://`:

| Input                 | Normalized                  |
| --------------------- | --------------------------- |
| `/run/secrets/token`  | `file:///run/secrets/token` |
| `file:///path`        | `file:///path`              |
| `https://example.com` | `https://example.com`       |

### Priority vs `ref` Tag

If a field has both `ref` and `refFrom` tags:

1. **`refFrom` takes precedence** if the referenced field has a value.
2. **`ref` is used as a fallback** if the referenced field is empty.

```go
// Tries to load from TokenPath first.
// If TokenPath is empty, falls back to loading from file:///run/secrets/token
Token string `ref:"file:///run/secrets/token" refFrom:"TokenPath"`
```

---

## Template Syntax

Both `ref` and `dsn` tags support a template syntax using `${...}` delimiters. Templates are processed using Go's `text/template` with custom delimiters.

### Expressions

| Syntax             | Description                         | Example                     |
| ------------------ | ----------------------------------- | --------------------------- |
| `${.FieldName}`    | Value of a field in the same struct | `${.Host}`, `${.Port}`      |
| `${.Nested.Field}` | Nested struct field access          | `${.Database.Host}`         |
| `${ref:uri}`       | Resolve a URI inline                | `${ref:file:///secret.txt}` |
| `${env:KEY}`       | Read an environment variable        | `${env:DB_USER}`            |

### Field Ordering Constraint

> **Important:** Fields referenced in templates must appear **earlier** in the struct definition. This is because fields are processed sequentially in declaration order.

```go
// ✅ Correct: SecretDir is defined before Password
type Config struct {
    SecretDir string `default:"/etc/secrets"`       // Field 0
    Password  string `ref:"file://${.SecretDir}/pass"` // Field 1 - can see Field 0
}

// ❌ Wrong: Password defined before SecretDir
type Config struct {
    Password  string `ref:"file://${.SecretDir}/pass"` // Field 0 - SecretDir is empty!
    SecretDir string `default:"/etc/secrets"`       // Field 1
}
```

### Inline URI Resolution with `${ref:uri}`

Resolve external URIs inline without storing them in fields:

```go
// Docker secrets
DSN string `dsn:"postgres://admin:${ref:file:///run/secrets/db_password}@db:5432/app"`

// Vault secrets
DSN string `dsn:"postgres://${ref:vault:///secret/db#user}:${ref:vault:///secret/db#pass}@localhost/db"`
```

### Inline Environment Variables with `${env:KEY}`

Read environment variables inline:

```go
DSN string `dsn:"postgres://${env:DB_USER}:${env:DB_PASSWORD}@${.Host}:5432/app"`
```

Environment variables respect the `WithEnvPrefix()` setting:

```go
// With WithEnvPrefix("APP_"):
// ${env:DB_USER} reads APP_DB_USER
```

### Strict Mode (DSN only)

By default, empty values produce empty strings. Enable strict mode to error on missing values:

```go
DSN string `dsn:"postgres://${.User}@${.Host}:5432/db" dsnStrict:"true"`
```

---

## `dsn` Tag

Composes connection strings from other fields using [template syntax](#template-syntax).

The `dsn` tag is processed **after** all other tags (`env`, `ref`, `default`), so referenced fields have their final values.

### Basic Example

```go
type Config struct {
    DBHost     string `yaml:"host" default:"localhost"`
    DBPort     int    `yaml:"port" default:"5432"`
    DBUser     string `yaml:"user"`
    DBPassword string `yaml:"password"`

    PostgresDSN string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:${.DBPort}/mydb"`
}
```

### Nested Struct Fields

```go
type Config struct {
    Database struct {
        Host string `yaml:"host" default:"localhost"`
        User string `yaml:"user"`
        Pass string `yaml:"password"`
    } `yaml:"database"`

    PostgresDSN string `dsn:"postgres://${.Database.User}:${.Database.Pass}@${.Database.Host}:5432/app"`
}
```

### Mixed Sources

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
