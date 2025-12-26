# Scanner Interface Example

Demonstrates implementing the `Scanner` interface for custom type conversion.

## Features

- Parse strings into custom enum types
- Support aliases (e.g., "pg" → "postgres")
- Works with `default` tag values

## The Scanner Interface

```go
type Scanner interface {
    Scan(src any) error
}
```

## Run

```bash
go run main.go
```
