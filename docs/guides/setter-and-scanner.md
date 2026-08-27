# Add dynamic defaults and custom conversions

Use a `Setter` when one value depends on other loaded values.
Use a `Scanner` when Fuda must convert text into your own Go type.

## Compute a value with Setter

```go
type Config struct {
    Host    string `yaml:"host" default:"localhost"`
    Port    int    `yaml:"port" default:"8080"`
    BaseURL string `yaml:"base_url"`
}

func (c *Config) SetDefaults() {
    if c.BaseURL == "" {
        c.BaseURL = fmt.Sprintf("http://%s:%d", c.Host, c.Port)
    }
}
```

Fuda calls `SetDefaults()` after it processes field tags.
The method can use the final `Host` and `Port` values.

## Convert text with Scanner

```go
type LogLevel int

func (level *LogLevel) Scan(src any) error {
    value, ok := src.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", src)
    }

    switch value {
    case "debug":
        *level = 0
    case "info":
        *level = 1
    default:
        return fmt.Errorf("unknown log level: %s", value)
    }

    return nil
}

type Config struct {
    LogLevel LogLevel `default:"info"`
}
```

Fuda calls `Scan(any) error` when it converts a tag value for the custom field.

## Know the order

Fuda processes environment, reference, and default tags before it calls a `Setter`.
It validates the final struct after `Setter` methods run.

## Next step

[Review configuration precedence](../getting-started/choosing-a-value.md) when you combine several features.
