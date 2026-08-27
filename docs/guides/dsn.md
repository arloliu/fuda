# Compose a DSN

Use `dsn` to build a connection string after Fuda has processed ordinary fields, environment values, references, and defaults.

```go
type Config struct {
    Host     string `yaml:"host" default:"localhost"`
    Port     int    `yaml:"port" default:"5432"`
    User     string `env:"DB_USER"`
    Password string `ref:"env://DB_PASSWORD"`
    DSN      string `dsn:"postgres://${.User}:${.Password}@${.Host}:${.Port}/app" dsnStrict:"true"`
}
```

Keep `Host`, `Port`, `User`, and `Password` before `DSN` in the struct.
Fuda then has their final values when it expands `${.Field}` expressions.

## Use URI templates

Use an environment value or reference inside a DSN template when it does not need its own struct field:

```go
RedisDSN string `dsn:"redis://:${env:REDIS_PASSWORD}@${.Host}:6379/0"`
```

```go
MongoDSN string `dsn:"mongodb://admin:${ref:file:///run/secrets/mongo_password}@${.Host}:27017/app"`
```

Use `WithTimeout` when an inline reference reaches an HTTP endpoint.

## Fail on missing fields

By default, a missing DSN expression becomes an empty string.
Add `dsnStrict:"true"` to return an error instead.
This protects connection strings that require every field.

## Next step

[Render a configuration template](templates.md) before Fuda decodes it.
