# Build your first configuration

## What you will learn

You will load a YAML file into a Go struct.
The same program will use a default port and let `APP_PORT` override it.

## Before you start

Finish the [installation step](install.md).
Create an empty Go package with a `main.go` file and a `config.yaml` file.

## Step 1: Define the configuration struct

Put this in `main.go`:

```go
package main

import (
    "fmt"
    "log"

    "github.com/arloliu/fuda"
)

type Config struct {
    Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
    Port int    `yaml:"port" default:"8080" env:"APP_PORT"`
}

func main() {
    loader, err := fuda.New().FromFile("config.yaml").Build()
    if err != nil {
        log.Fatal(err)
    }

    var cfg Config
    if err := loader.Load(&cfg); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
}
```

## Step 2: Add a YAML file

Put this in `config.yaml`:

```yaml
host: 0.0.0.0
```

The file sets `Host`.
It leaves `Port` out, so Fuda uses the `default:"8080"` tag.

## Step 3: Run the program

Run:

```bash
go run .
```

You should see:

```text
Server: 0.0.0.0:8080
```

## Step 4: Override the port with an environment variable

Run the same program with `APP_PORT`:

```bash
APP_PORT=9090 go run .
```

You should see:

```text
Server: 0.0.0.0:9090
```

## What happened

Fuda reads `Host` from YAML.
No YAML value exists for `Port`, so it starts with `8080` from the `default` tag.
When `APP_PORT` exists, the `env` tag replaces the file or default value.

## Common mistakes

Use the field names in the `yaml` tags as keys in `config.yaml`.
Set `APP_PORT` in the same command that runs the program.
On Windows PowerShell, use `$env:APP_PORT = "9090"` before `go run .`.

## Next step

[Choose a value](choosing-a-value.md) to see the precedence rule in one place.
