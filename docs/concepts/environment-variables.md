# Environment variables

Use an `env` tag when an environment variable should override a file or default value.

```go
type Config struct {
    Port int `yaml:"port" default:"8080" env:"APP_PORT"`
}
```

With this field, `APP_PORT` has priority over `port` in YAML and over `default:"8080"`.

```bash
APP_PORT=9090 go run .
```

Fuda converts the string `9090` into an integer before it stores the value in `Port`.

## Add a prefix

Use `WithEnvPrefix` when your struct tags should contain short, reusable names.

```go
type Config struct {
    Port int `yaml:"port" default:"8080" env:"PORT"`
}

loader, err := fuda.New().
    FromFile("config.yaml").
    WithEnvPrefix("APP_").
    Build()
```

Fuda now looks up `APP_PORT`.
It joins the prefix and the `env` tag exactly, so include the trailing underscore in `APP_` when you want one.

## Empty values

Fuda treats a set environment variable as an override.
For string fields, an empty value overrides a non-empty default.
Unset the variable when you want to use the YAML or default value again.

## Next step

[Validate the final configuration](validation.md) before you start your application.
