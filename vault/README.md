# Fuda Vault resolver

Load Vault secret fields through `vault://` references.

```bash
go get github.com/arloliu/fuda/vault
```

```go
resolver, err := vault.NewResolver(
    vault.WithAddress("https://vault.example.com:8200"),
    vault.WithToken(os.Getenv("VAULT_TOKEN")),
)
```

Read the [Vault integration guide](../docs/integrations/vault.md) for URI syntax, authentication options, and safe secret handling.
