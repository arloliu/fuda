# Apply programmatic overrides

Use `WithOverrides` when your program must change configuration after reading a file.

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithOverrides(map[string]any{
        "server.port": 9090,
    }).
    Build()
```

This example changes the nested YAML value at `server.port` before Fuda decodes the source into your struct.

## Know the order

Fuda processes template content first.
It applies overrides to the rendered source before YAML decoding.
After decoding, an `env` tag can still replace the overridden field.

```go
type Config struct {
    Port int `yaml:"port" env:"APP_PORT"`
}
```

With `APP_PORT=8080`, the environment value wins over an override that set `port` to `9090`.

Use overrides for values your program controls, such as a test port or a command-line option.
Keep operator-controlled configuration in files or environment variables.

## Next step

[Add dynamic defaults or custom conversions](setter-and-scanner.md).
