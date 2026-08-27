# Configuration files

Fuda reads YAML or JSON from a file, reader, or byte slice.
Use a `yaml` tag when the file key should differ from the Go field name.

```go
type Config struct {
    Host string `yaml:"host" default:"localhost"`
    Port int    `yaml:"port" default:"8080"`
}
```

```yaml
host: 0.0.0.0
port: 3000
```

Load the file with the builder:

```go
loader, err := fuda.New().FromFile("config.yaml").Build()
if err != nil {
    return err
}

var cfg Config
if err := loader.Load(&cfg); err != nil {
    return err
}
```

Fuda matches YAML keys to the `yaml` tags and decodes them into the struct.
An ordinary file value wins over the field's `default` value.

## Missing and explicit values

When `port` is missing, the `default:"8080"` tag supplies the value.
An explicit zero value remains an explicit choice:

```yaml
port: 0
```

Fuda keeps `0` instead of replacing it with `8080`.
The same rule applies to `false` and an empty string.

YAML `null` counts as absent for default handling.
With `port: null`, Fuda applies `default:"8080"`.

## Next step

[Use defaults](defaults.md) to give missing values a useful starting point.
