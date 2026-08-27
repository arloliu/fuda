# Load a dotenv file

Use dotenv support when local development needs environment values from a file.
Fuda loads dotenv files before it resolves `env` tags.

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithDotEnv(".env").
    Build()
```

With an `env:"APP_PORT"` tag, a value from `.env` can populate `APP_PORT`.
Fuda ignores missing dotenv files, so optional local files do not block startup.

## Use overlays

Load files in order when you need a base file and a local overlay:

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithDotEnvFiles([]string{".env", ".env.local"}).
    Build()
```

Later files can add values after earlier files.
Existing process environment values still win by default.

## Search for a dotenv file

Search a few known locations when the working directory can vary:

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithDotEnvSearch(".env", []string{".", "./config", "/etc/myapp"}).
    Build()
```

Fuda uses the first matching file.

## Let dotenv replace process values

Use `DotEnvOverride()` only when the dotenv file should replace values that already exist in the process environment:

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithDotEnv(".env", fuda.DotEnvOverride()).
    Build()
```

Keep secrets out of committed dotenv files.
Use your deployment environment or a secret manager for production values.

## Next step

[Load values from external references](references.md).
