# Validation

Fuda uses `go-playground/validator` tags when it loads configuration through `fuda.New()` or a convenience loader such as `fuda.LoadBytes`.

```go
type Config struct {
    Host string `yaml:"host" validate:"required"`
    Port int    `yaml:"port" validate:"min=1,max=65535"`
}
```

## Load a valid configuration

```go
var cfg Config
err := fuda.LoadBytes([]byte("host: localhost\nport: 8080\n"), &cfg)
if err != nil {
    return err
}
```

Fuda loads the fields, applies tags, then validates the final struct.
This order lets a default or environment variable satisfy a validation rule.

## Handle a validation failure

```go
var cfg Config
err := fuda.LoadBytes([]byte("host: \"\"\nport: 0\n"), &cfg)
if err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}
```

The `required` and `min` rules fail in this example.
Return the error during startup so your application does not run with invalid configuration.

## Next step

[Choose the right data type](data-types.md) for durations and byte sizes.
