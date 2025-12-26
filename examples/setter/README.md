# Setter Interface Example

Demonstrates implementing the `Setter` interface for dynamic defaults that can't be expressed as static tag values.

## Features

- Generate random IDs at load time
- Get hostname from system
- Compute values from other fields
- Works with nested structs

## The Setter Interface

```go
type Setter interface {
    SetDefaults()
}
```

## Run

```bash
go run main.go
```
