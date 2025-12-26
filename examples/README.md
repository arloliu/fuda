# Fuda Examples

This directory contains examples demonstrating fuda's features.

## Examples

| Example | Description |
|---------|-------------|
| [basic](basic/) | Default values, environment overrides, builder pattern |
| [refs](refs/) | External references with `ref` and `refFrom` tags |
| [template](template/) | Go template processing for dynamic configs |
| [validation](validation/) | Struct validation with go-playground/validator |
| [scanner](scanner/) | Custom type conversion with Scanner interface |
| [setter](setter/) | Dynamic defaults with Setter interface |
| [watcher](watcher/) | Hot-reload configuration watching |

## Running Examples

```bash
cd examples/<name>
go run main.go
```

## Feature Matrix

| Feature | Tag/Package | Example |
|---------|-------------|---------|
| YAML/JSON parsing | `yaml`, `json` | basic |
| Default values | `default` | basic |
| Environment vars | `env` | basic |
| Static refs | `ref` | refs |
| Dynamic refs | `refFrom` | refs |
| Templates | `WithTemplate()` | template |
| Validation | `validate` | validation |
| Custom types | `Scanner` interface | scanner |
| Dynamic defaults | `Setter` interface | setter |
| Hot-reload | `fuda/watcher` | watcher |
