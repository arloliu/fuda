# Write a custom resolver

Implement `fuda.RefResolver` when your application needs a URI scheme that Fuda does not provide.

```go
type Resolver struct{}

func (Resolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
    if !strings.HasPrefix(uri, "config://") {
        return nil, fmt.Errorf("unsupported URI: %s", uri)
    }

    return []byte("value from the config service"), nil
}
```

Pass the resolver to the builder:

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithRefResolver(Resolver{}).
    Build()
```

Fuda calls `Resolve(context.Context, string) ([]byte, error)` for `ref` and `refFrom` values.
Respect the context so `WithTimeout` can limit network work.
Make the resolver safe for concurrent calls.

## Next step

[Review the reference tag syntax](../reference/tags.md).
