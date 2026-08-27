# Defaults

Use the `default` tag to make a field usable when no higher-priority source supplied it.

```go
type Config struct {
    Host    string `default:"localhost"`
    Port    int    `default:"8080"`
    Enabled bool   `default:"true"`
}
```

Fuda converts the tag value to the field's Go type.
For example, `default:"8080"` becomes an `int` for `Port`.

## Defaults fill missing values

This YAML file leaves `port` out:

```yaml
host: 0.0.0.0
```

The resulting configuration has `Host` set to `0.0.0.0` and `Port` set to `8080`.

Fuda does not replace a value that the file supplied.
These values stay exactly as written:

```yaml
port: 0
enabled: false
host: ""
```

Use `null` when you want a YAML key to behave as missing:

```yaml
port: null
```

Fuda then applies the default value for `Port`.

## When to use a default

Choose defaults that make local development straightforward.
Require an environment variable or validation rule for values that production must provide.

## Next step

[Override configuration with environment variables](environment-variables.md).
