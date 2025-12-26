# Custom Resolvers

This guide covers implementing custom `RefResolver` implementations for the fuda library.

## Overview

A `RefResolver` fetches content from URIs referenced in `ref` and `refFrom` tags. The library includes built-in resolvers for:
- `file://` - Local file system
- `http://` and `https://` - HTTP endpoints

## Interface

```go
type RefResolver interface {
    Resolve(ctx context.Context, uri string) ([]byte, error)
}
```

## Implementing a Custom Resolver

### Example: Vault Resolver

```go
package main

import (
    "context"
    "fmt"
    "net/url"

    vault "github.com/hashicorp/vault/api"
)

type VaultResolver struct {
    client *vault.Client
}

func NewVaultResolver(addr, token string) (*VaultResolver, error) {
    cfg := vault.DefaultConfig()
    cfg.Address = addr

    client, err := vault.NewClient(cfg)
    if err != nil {
        return nil, err
    }
    client.SetToken(token)

    return &VaultResolver{client: client}, nil
}

func (r *VaultResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
    u, err := url.Parse(uri)
    if err != nil {
        return nil, fmt.Errorf("invalid URI: %w", err)
    }

    if u.Scheme != "vault" {
        return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
    }

    // URI format: vault:///secret/data/myapp#field
    path := u.Path
    field := u.Fragment

    secret, err := r.client.Logical().ReadWithContext(ctx, path)
    if err != nil {
        return nil, fmt.Errorf("vault read failed: %w", err)
    }

    if secret == nil || secret.Data == nil {
        return nil, fmt.Errorf("secret not found: %s", path)
    }

    data, ok := secret.Data["data"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("unexpected secret format")
    }

    value, ok := data[field].(string)
    if !ok {
        return nil, fmt.Errorf("field not found: %s", field)
    }

    return []byte(value), nil
}
```

### Using with Fuda

```go
vaultResolver, _ := NewVaultResolver("https://vault.example.com", os.Getenv("VAULT_TOKEN"))

loader, _ := fuda.New().
    FromFile("config.yaml").
    WithRefResolver(vaultResolver).
    Build()
```

## Composite Resolver

To support multiple schemes, create a composite resolver:

```go
type CompositeResolver struct {
    resolvers map[string]fuda.RefResolver
}

func (c *CompositeResolver) Register(scheme string, r fuda.RefResolver) {
    c.resolvers[scheme] = r
}

func (c *CompositeResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
    u, _ := url.Parse(uri)
    resolver, ok := c.resolvers[u.Scheme]
    if !ok {
        return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
    }
    return resolver.Resolve(ctx, uri)
}
```

## Caching

For performance with repeated references, wrap your resolver with caching:

```go
type CachingResolver struct {
    inner fuda.RefResolver
    cache sync.Map
}

func (c *CachingResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
    if cached, ok := c.cache.Load(uri); ok {
        return cached.([]byte), nil
    }

    data, err := c.inner.Resolve(ctx, uri)
    if err != nil {
        return nil, err
    }

    c.cache.Store(uri, data)
    return data, nil
}
```

## Best Practices

1. **Respect context** - Always check `ctx.Done()` for cancellation
2. **Wrap errors** - Use `fmt.Errorf("...: %w", err)` for error chains
3. **Validate URIs** - Check scheme before processing
4. **Handle timeouts** - The caller sets timeout via `WithTimeout()`
