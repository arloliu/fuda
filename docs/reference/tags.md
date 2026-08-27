# Tag reference

Attach Fuda tags to exported struct fields.

| Tag | Example | Purpose |
| --- | --- | --- |
| `yaml` | `yaml:"host"` | Maps a YAML key to the field. |
| `json` | `json:"host"` | Maps a JSON key to the field. |
| `default` | `default:"8080"` | Supplies a value when no source supplied one. |
| `env` | `env:"APP_PORT"` | Reads an environment override. |
| `ref` | `ref:"file:///run/secrets/token"` | Resolves a fixed external URI. |
| `refFrom` | `refFrom:"TokenURI"` | Resolves the URI stored in an earlier field. |
| `dsn` | `dsn:"postgres://${.Host}:5432/app"` | Builds a string from fields, env values, or references. |
| `dsnStrict` | `dsnStrict:"true"` | Makes a DSN fail when an expression is empty. |
| `validate` | `validate:"required"` | Applies a validator rule after loading. |

Tags can work together:

```go
type Config struct {
    Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
    Port int    `yaml:"port" default:"8080" env:"APP_PORT" validate:"min=1,max=65535"`
}
```

For the basic value rule, read [Choose a value](../getting-started/choosing-a-value.md).
For external values, read [Load external references](../guides/references.md).
