# Fuda

<p align="center">
  <img src="docs/assets/images/fuda-logo.png" alt="Fuda logo" width="240">
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/arloliu/fuda.svg)](https://pkg.go.dev/github.com/arloliu/fuda)

Fuda is a struct-tag-first configuration library for Go.
It reads configuration files, supplies defaults, accepts environment overrides, resolves external values, and validates the final struct.

Read the [documentation site](https://arloliu.github.io/fuda/) for the complete beginner-first guide,
or the [package documentation](https://pkg.go.dev/github.com/arloliu/fuda) for the API reference.

## Install

```bash
go get github.com/arloliu/fuda
```

## Quick example

```go
type Config struct {
    Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
    Port int    `yaml:"port" default:"8080" env:"APP_PORT"`
}

var cfg Config
if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
    log.Fatal(err)
}
```

```yaml
host: 0.0.0.0
```

Fuda uses `0.0.0.0` for `Host` and `8080` for `Port`.
`APP_PORT=9090` overrides the default or YAML value for `Port`.

## Features

- YAML and JSON configuration files
- `default` and `env` tags
- Dotenv, templates, and programmatic overrides
- `ref`, `refFrom`, DSN composition, and Vault support
- Validation, dynamic defaults, and custom Scanner types
- File and remote-source watching

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
