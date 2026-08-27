# Choose a value

## What you will learn

You will predict the value Fuda assigns when an environment variable, a configuration file, and a default tag all provide candidates.

## Before you start

Read [your first configuration](first-configuration.md).
Use this field as the example:

```go
Port int `yaml:"port" default:"8080" env:"APP_PORT"`
```

## Step 1: Start with the default

When neither YAML nor `APP_PORT` supplies a port, Fuda uses `8080`.

```yaml
# config.yaml has no port key
```

The resulting value is `8080`.

## Step 2: Add a configuration-file value

Add a YAML value:

```yaml
port: 3000
```

The resulting value is `3000`.
The file value replaces the default.

## Step 3: Add an environment override

Set the environment variable:

```bash
APP_PORT=9090 go run .
```

The resulting value is `9090`.
The environment variable replaces the file value.

## What happened

Fuda checks values in this order:

1. Environment variable
2. Configuration file
3. Default tag

The first available value wins.

Read more about [configuration files](../concepts/configuration-files.md) and [defaults](../concepts/defaults.md).
Then read about [environment variables](../concepts/environment-variables.md).

## Common mistakes

An empty environment variable is still an environment value.
Use `unset APP_PORT` on macOS or Linux when you want to test the YAML value again.
Use the exact variable name from the `env` tag unless you configure an environment prefix.

## Next step

[Run the first configuration again](first-configuration.md) and change one input at a time.
