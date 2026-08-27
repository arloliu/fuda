# Load HashiCorp Vault secrets

The Vault resolver lives in the separate `github.com/arloliu/fuda/vault` module.

```bash
go get github.com/arloliu/fuda/vault
```

Create a resolver with a Vault address and an authentication option:

```go
resolver, err := vault.NewResolver(
    vault.WithAddress("https://vault.example.com:8200"),
    vault.WithToken(os.Getenv("VAULT_TOKEN")),
)
if err != nil {
    return err
}
```

Give Fuda the resolver:

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithRefResolver(resolver).
    Build()
```

Use a Vault URI in a `ref` tag:

```go
DBPassword string `ref:"vault:///secret/data/myapp#password"`
```

The URI shape is `vault:///<mount>/<path>#<field>`.
Use a placeholder path in documentation and keep the address, token, and secret values outside source control.

The resolver also supports Kubernetes and AppRole authentication through `WithKubernetesAuth` and `WithAppRole`.

## Next step

[Write a custom resolver](custom-resolvers.md) for another secret backend.
