# Configuration precedence

Run a small program that loads YAML, uses a default, and accepts an environment override.

## Run

```bash
go run .
```

It prints `Server: 0.0.0.0:8080`.

## Override a value

```bash
APP_PORT=9090 go run .
```

It prints `Server: 0.0.0.0:9090`.

Read the [first configuration guide](../../docs/getting-started/first-configuration.md) for the walkthrough.
