# Validation Example

Demonstrates using the `validate` tag with go-playground/validator rules.

## Features

- Required fields
- Numeric constraints (min, max)
- String enums (oneof)
- URL and email validation
- Nested struct validation

## Run

```bash
go run main.go
```

## Common Validation Rules

| Rule | Description |
|------|-------------|
| `required` | Field must not be zero value |
| `min=N` | Minimum value/length |
| `max=N` | Maximum value/length |
| `oneof=a b c` | Must be one of listed values |
| `url` | Must be valid URL |
| `email` | Must be valid email |
| `hostname` | Must be valid hostname |
| `ip` | Must be valid IP address |
