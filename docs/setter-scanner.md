# Setter and Scanner Interfaces

This guide covers the `Setter` and `Scanner` interfaces for customizing configuration behavior.

## Setter Interface

The `Setter` interface enables dynamic defaults that can't be expressed as static tag values.

### Interface

```go
type Setter interface {
    SetDefaults()
}
```

### When It's Called

`SetDefaults()` is called **after** all tag processing completes:

1. YAML/JSON unmarshaling
2. Environment variable overrides (`env` tag)
3. Reference resolution (`ref`/`refFrom` tags)
4. Static defaults (`default` tag)
5. **→ SetDefaults() called**
6. Validation (`validate` tag)

### Example: Dynamic UUID

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

### Example: Computed Defaults

```go
type ServerConfig struct {
    Host        string `default:"localhost"`
    Port        int    `default:"8080"`
    BaseURL     string // Computed from Host + Port
    MaxBodySize int64  `default:"1048576"` // 1MB
}

func (c *ServerConfig) SetDefaults() {
    if c.BaseURL == "" {
        c.BaseURL = fmt.Sprintf("http://%s:%d", c.Host, c.Port)
    }
}
```

### Nested Structs

`SetDefaults` is called in post-order traversal (children before parents):

```go
type App struct {
    DB     Database
    Server Server
}

type Database struct {
    Pool int `default:"10"`
}

func (d *Database) SetDefaults() {
    // Called first
}

type Server struct {
    Port int `default:"8080"`
}

func (s *Server) SetDefaults() {
    // Called second
}

func (a *App) SetDefaults() {
    // Called last - can reference child values
}
```

---

## Scanner Interface

The `Scanner` interface enables custom string-to-value conversion for the `default` tag.

### Interface

```go
type Scanner interface {
    Scan(src any) error
}
```

### When It's Called

When processing a `default` tag, fuda checks if the target type implements `Scanner`. If so, `Scan()` is called with the default string value instead of using built-in conversion.

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
    case "debug":
        *l = Debug
    case "info":
        *l = Info
    case "warn":
        *l = Warn
    case "error":
        *l = Error
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

### Example: Custom Duration

```go
type Duration struct {
    time.Duration
}

func (d *Duration) Scan(src any) error {
    s, ok := src.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", src)
    }

    // Support "1d" for days (not in stdlib)
    if strings.HasSuffix(s, "d") {
        days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
        if err != nil {
            return err
        }
        d.Duration = time.Duration(days) * 24 * time.Hour
        return nil
    }

    var err error
    d.Duration, err = time.ParseDuration(s)
    return err
}

// Usage
type Config struct {
    CacheTTL Duration `default:"7d"`
}
```

### Example: String Set

```go
type StringSet map[string]struct{}

func (s *StringSet) Scan(src any) error {
    str, ok := src.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", src)
    }

    *s = make(StringSet)
    for _, item := range strings.Split(str, ",") {
        (*s)[strings.TrimSpace(item)] = struct{}{}
    }
    return nil
}

// Usage
type Config struct {
    AllowedHosts StringSet `default:"localhost,127.0.0.1"`
}
```

---

## Combining Setter and Scanner

Both interfaces can be used together for maximum flexibility:

```go
type Config struct {
    Mode      RunMode   `yaml:"mode" default:"production"`
    StartTime time.Time // Dynamic default
}

type RunMode string

func (r *RunMode) Scan(src any) error {
    s, ok := src.(string)
    if !ok {
        return fmt.Errorf("expected string")
    }

    switch strings.ToLower(s) {
    case "dev", "development":
        *r = "development"
    case "prod", "production":
        *r = "production"
    default:
        return fmt.Errorf("unknown mode: %s", s)
    }
    return nil
}

func (c *Config) SetDefaults() {
    if c.StartTime.IsZero() {
        c.StartTime = time.Now()
    }
}
```

## Best Practices

| Practice | Details |
|----------|---------|
| **Check zero values** | In `SetDefaults()`, always check if a value is zero before setting |
| **Handle type assertions** | In `Scan()`, validate the input type before processing |
| **Use pointer receivers** | Both interfaces must use pointer receivers to modify the value |
| **Keep idempotent** | `SetDefaults()` should be safe to call multiple times |
