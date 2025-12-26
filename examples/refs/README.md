# External References Example

Demonstrates loading secrets from external files using `ref` and `refFrom` tags.

## Features

- `ref` - Load from a fixed URI (file://, http://, https://)
- `refFrom` - Load from a URI stored in another config field

## Run

```bash
go run main.go
```

## Use Case

Perfect for Kubernetes secrets mounted as files:

```yaml
db_password_path: "file:///run/secrets/db_password"
```
