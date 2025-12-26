# Template Processing Example

Demonstrates using Go template syntax to dynamically generate configuration.

## Features

- Conditional logic with `{{ if }}` / `{{ else }}`
- Variable substitution with `{{ .Variable }}`
- Works before YAML parsing

## Run

```bash
# Development (default)
go run main.go

# Production
ENVIRONMENT=prod REGION=us-east-1 go run main.go

# Staging
ENVIRONMENT=staging go run main.go
```
