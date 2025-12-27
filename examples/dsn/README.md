# DSN Composition Example

This example demonstrates using the `dsn` tag to compose connection strings from multiple configuration sources.

## Three Patterns Demonstrated

### Pattern 1: Field References `${.FieldName}`

Compose DSN from other config fields that are populated via YAML, env vars, or defaults:

```go
type Config struct {
    DBHost     string `yaml:"db_host" default:"localhost"`
    DBUser     string `env:"DB_USER"`
    DBPassword string `env:"DB_PASSWORD"`

    PostgresDSN string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:5432/app"`
}
```

### Pattern 2: Inline Environment Variables `${env:KEY}`

Read environment variables directly in the DSN template:

```go
RedisDSN string `dsn:"redis://:${env:REDIS_PASSWORD}@${.RedisHost}:${.RedisPort}/0"`
```

### Pattern 3: Inline File Reference `${ref:uri}`

Resolve secrets from files or other sources inline:

```go
// From file (use absolute path with file://)
MongoDSN string `dsn:"mongodb://admin:${ref:file:///run/secrets/mongo_pass}@host:27017/db"`

// From vault (requires vault resolver)
DSN string `dsn:"postgres://${ref:vault:///secret/db#user}:${ref:vault:///secret/db#pass}@host:5432/db"`
```

## Running the Example

```bash
# Run the example
go run main.go
```

## Expected Output

```
=== DSN Composition Example ===

--- Pattern 1: DSN from Field References ---
Config struct uses: `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:${.DBPort}/${.DBName}"`
  DB Host:       db.example.com (from yaml)
  DB User:       postgres_admin (from env:DB_USER)
  PostgreSQL DSN: postgres://postgres_admin:****@db.example.com:5432/production

--- Pattern 2: DSN with Inline ${env:KEY} ---
Config struct uses: `dsn:"redis://:${env:REDIS_PASSWORD}@${.RedisHost}:${.RedisPort}/0"`
  Redis Host:    redis.example.com
  Redis DSN:     redis://:****@redis.example.com:6380/0

--- Pattern 3: DSN with Inline ${ref:file://...} ---
  Secret file:   /tmp/fuda-example/mongo_password.txt
  DSN template:  `dsn:"mongodb://admin:${ref:file:///path/to/secret}@host:27017/db"`
  Would resolve: mongodb://admin:mongo_secret_123@localhost:27017/db
```

## Summary

| Syntax | Description | Example |
|--------|-------------|---------|
| `${.Field}` | Reference struct field | `${.DBHost}` |
| `${.Nested.Field}` | Reference nested field | `${.Database.Host}` |
| `${env:KEY}` | Read environment variable | `${env:DB_PASSWORD}` |
| `${ref:uri}` | Resolve from file/vault/http | `${ref:vault:///secret#pass}` |
