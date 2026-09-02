# Fuda

<p align="center">
  <img src="docs/assets/images/fuda-logo.png" alt="Fuda logo" width="240">
</p>

[![CI](https://github.com/arloliu/fuda/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/arloliu/fuda/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/arloliu/fuda.svg)](https://pkg.go.dev/github.com/arloliu/fuda)
[![Go Report Card](https://goreportcard.com/badge/github.com/arloliu/fuda)](https://goreportcard.com/report/github.com/arloliu/fuda)
[![License](https://img.shields.io/github/license/arloliu/fuda)](LICENSE)

**Struct tags in, a validated config out.**
Fuda reads your YAML or JSON file, layers in defaults and environment overrides, resolves secrets and connection strings, and validates the result.
All of it is declared on the struct itself, with no glue code in `main()`.

Read the [documentation site](https://arloliu.github.io/fuda/) for the complete beginner-first guide,
or the [package documentation](https://pkg.go.dev/github.com/arloliu/fuda) for the API reference.

## Why Fuda

- **One struct, one source of truth.**
  Keys, defaults, env vars, secrets, and validation rules live next to the field they describe.
  Nothing is spread across a config struct, a flags package, and a validator call.
- **Secrets without a secrets SDK.**
  Pull values from files, HTTP endpoints, or Vault straight into a field with `ref`/`refFrom`.
  Compose connection strings from those values with `dsn`, with no manual string building.
- **Human input, typed output.**
  Write `10MiB` or `7d` in config and get an `int64` or `time.Duration` back.
  Write `debug` and get your own enum type back via the `Scanner` interface.
- **Fails loudly, not late.**
  `validate` tags check the fully loaded struct before your program starts.
  You get field-level errors instead of a nil pointer three hours into a shift.
- **Grows with you.**
  Start with `fuda.LoadFile`, then reach for the builder for env prefixes, dotenv overlays, templated config, or live file/remote watching.
  None of that requires changing the struct you already wrote.

## Install

```bash
go get github.com/arloliu/fuda
```

## Quick example

One `Load` call resolves a secret, composes a DSN, validates the struct, and sets a dynamic default.

```go
// LogLevel is a custom type; Scan lets fuda convert the default/YAML
// string into it via the Scanner interface.
type LogLevel int

const (
    LevelInfo LogLevel = iota
    LevelDebug
)

func (l *LogLevel) Scan(src any) error {
    if src == "debug" {
        *l = LevelDebug
    }
    return nil
}

type Config struct {
    AppName  string   `yaml:"app_name" validate:"required"`
    Env      string   `yaml:"env" default:"dev" validate:"oneof=dev staging prod"`
    LogLevel LogLevel `yaml:"log_level" default:"info"`

    Host string `yaml:"host" default:"0.0.0.0" env:"APP_HOST"`
    Port int    `yaml:"port" default:"8080" env:"APP_PORT" validate:"min=1,max=65535"`

    DBUser     string `yaml:"db_user" default:"app"`
    DBPassword string `ref:"file://secrets/db_password.txt"`
    DBHost     string `yaml:"db_host" default:"localhost"`
    DBName     string `yaml:"db_name" default:"orders"`
    DSN        string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}/${.DBName}"`

    StartedAt time.Time
}

// SetDefaults runs after tags are applied, for defaults tags can't express.
func (c *Config) SetDefaults() {
    c.StartedAt = time.Now()
}

func main() {
    var cfg Config
    if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
        var verr *fuda.ValidationError
        if errors.As(err, &verr) {
            log.Fatalf("invalid config: %v", verr.Errors)
        }
        log.Fatal(err)
    }

    fmt.Printf("%s DSN: %s\n", cfg.AppName, cfg.DSN)
}
```

Pair it with a minimal `config.yaml`:

```yaml
app_name: orders-api
db_user: app
```

With `secrets/db_password.txt` holding the database password, Fuda resolves the ref, composes `DSN`, applies every default, validates the result, and stamps `StartedAt` before your program ever sees `cfg`.
`APP_PORT=9090` overrides `Port` at runtime without touching the file.

## Feature tour

The struct above already covers files, defaults, env overrides, secrets, DSN composition, validation, dynamic defaults, and custom types.
Fuda also handles the rest of a real service's configuration.

### Human-readable sizes and durations

```go
type Config struct {
    Timeout   time.Duration `yaml:"timeout"`
    Retention fuda.Duration `yaml:"retention"`
    CacheSize fuda.ByteSize `yaml:"cache_size"`
}
```

```yaml
timeout: 30s
retention: 7d
cache_size: 10MiB
```

Fuda parses `7d` into a duration and `10MiB` into a byte count, no manual parsing required.
`fuda.Duration` and `fuda.ByteSize` also marshal back to a readable string instead of raw nanoseconds or bytes.

### Beyond a single file

- **Environment prefixes.**
  `WithEnvPrefix("APP_")` matches `env` tags against `APP_HOST` instead of `HOST`.
- **Dotenv overlays.**
  `WithDotEnvFiles([]string{".env", ".env.local"})` layers environment files before Fuda reads the struct.
- **Templated config.**
  `WithTemplate(data)` renders the YAML or JSON file as a Go template before parsing it.
- **Live reload.**
  `fuda/watcher` reloads and revalidates the struct when a watched file or remote reference changes, and delivers the new value on a channel.
- **Vault secrets.**
  `fuda/vault` resolves `ref:"vault:///secret/data/app#password"` against a running Vault server, with Kubernetes and AppRole auth built in.

Each of these is a runnable program under [examples/](examples/README.md).

## Tag reference

| Tag | Example | What it does |
| --- | --- | --- |
| `yaml` / `json` | `yaml:"host"` | Maps a file key to the field. |
| `default` | `default:"8080"` | Supplies a value when no other source did. |
| `env` | `env:"APP_PORT"` | Reads an environment variable override. |
| `ref` | `ref:"file:///run/secrets/token"` | Resolves a fixed external URI (file, HTTP, Vault). |
| `refFrom` | `refFrom:"TokenURI"` | Resolves the URI stored in another field. |
| `dsn` | `dsn:"postgres://${.Host}:5432/app"` | Builds a string from fields, env values, or refs. |
| `validate` | `validate:"required,min=1"` | Applies a validator rule after loading. |

See the [full tag reference](https://arloliu.github.io/fuda/reference/tags/) for `dsnStrict` and every validator rule.

## Learn more

- [Install Fuda](https://arloliu.github.io/fuda/getting-started/install/)
- [Build your first configuration](https://arloliu.github.io/fuda/getting-started/first-configuration/)
- [Tag reference](https://arloliu.github.io/fuda/reference/tags/)
- [Package documentation](https://pkg.go.dev/github.com/arloliu/fuda)
- [Examples](examples/README.md)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.

## License

Fuda is licensed under the [Apache License 2.0](LICENSE).
