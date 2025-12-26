# Vault Resolver

The `fuda/vault` package provides a HashiCorp Vault resolver for fetching secrets directly into your configuration struct.

## Installation

The Vault package is a **separate Go module** to avoid adding the Vault SDK as a core fuda dependency. Install it with:

```bash
go get github.com/arloliu/fuda/vault
```

Then import:

```go
import "github.com/arloliu/fuda/vault"
```

## Quick Start

```go
package main

import (
    "log"
    "os"

    "github.com/arloliu/fuda"
    "github.com/arloliu/fuda/vault"
)

type Config struct {
    DBPassword string `ref:"vault:///secret/data/myapp#db_password"`
    APIKey     string `ref:"vault:///secret/data/myapp#api_key"`
}

func main() {
    // Create Vault resolver
    resolver, err := vault.NewResolver(
        vault.WithAddress("https://vault.example.com:8200"),
        vault.WithToken(os.Getenv("VAULT_TOKEN")),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use with fuda
    loader, err := fuda.New().
        FromFile("config.yaml").
        WithRefResolver(resolver).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    var cfg Config
    if err := loader.Load(&cfg); err != nil {
        log.Fatal(err)
    }
}
```

## URI Format

```
vault:///<mount>/<path>#<field>
```

| Component | Description |
|-----------|-------------|
| `mount` | Secrets engine mount path (e.g., `secret`) |
| `path` | Path to the secret |
| `field` | Field name within the secret |

### Examples

```go
// KV v2 (versioned secrets)
DBPassword string `ref:"vault:///secret/data/myapp#password"`

// KV v1 (unversioned)
APIKey string `ref:"vault:///kv/myapp#api_key"`

// Database dynamic secrets
DBUser string `ref:"vault:///database/creds/readonly#username"`

// Dynamic path from config
SecretPath string `yaml:"secret_path"`
Token      string `refFrom:"SecretPath"`  // Supports vault:// URIs
```

## Authentication Methods

### Token Authentication

Simplest method, suitable for development or CI/CD:

```go
vault.WithToken(os.Getenv("VAULT_TOKEN"))
```

### Kubernetes Authentication

Recommended for pods running in Kubernetes:

```go
vault.WithKubernetesAuth(
    "my-app-role",
    "/var/run/secrets/kubernetes.io/serviceaccount/token",
)
```

For custom auth mount paths:

```go
vault.WithKubernetesAuthMount(
    "custom-k8s",  // Mount path
    "my-role",
    "/var/run/secrets/kubernetes.io/serviceaccount/token",
)
```

### AppRole Authentication

Machine-to-machine authentication:

```go
vault.WithAppRole(
    os.Getenv("VAULT_ROLE_ID"),
    os.Getenv("VAULT_SECRET_ID"),
)
```

## Additional Options

```go
// Vault Enterprise namespace
vault.WithNamespace("my-team")

// Custom TLS configuration
vault.WithTLSConfig(&api.TLSConfig{
    CACert:   "/path/to/ca.crt",
    Insecure: false,
})
```

## Kubernetes Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      serviceAccountName: my-app
      containers:
        - name: app
          image: my-app:latest
          volumeMounts:
            - name: config
              mountPath: /etc/config
      volumes:
        - name: config
          configMap:
            name: my-app-config
```

Create a Vault role for the service account:

```bash
vault write auth/kubernetes/role/my-app-role \
    bound_service_account_names=my-app \
    bound_service_account_namespaces=default \
    policies=my-app-policy \
    ttl=1h
```

## Thread Safety

The `Resolver` is safe for concurrent use after creation. Multiple goroutines can call `Resolve()` simultaneously.

## Error Handling

```go
_, err := resolver.Resolve(ctx, "vault:///secret/data/missing#field")
if err != nil {
    // Common errors:
    // - "vault secret not found at ..."
    // - "field not found in vault secret at ..."
    // - "vault authentication failed: ..."
}
```
