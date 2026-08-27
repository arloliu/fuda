# Load external references

Use `ref` to load a field from a URI.
Use `refFrom` when another field contains the URI.

```go
type Config struct {
    PasswordPath string `yaml:"password_path"`
    Password     string `refFrom:"PasswordPath" ref:"env://APP_PASSWORD"`
}
```

Fuda tries `PasswordPath` first.
When that field has no URI, it falls back to the `ref` tag.

```yaml
password_path: file:///run/secrets/db_password
```

The default resolver supports these URI schemes:

| Scheme | Example | Use |
| --- | --- | --- |
| `file://` | `file:///run/secrets/db_password` | Read a mounted secret or local file. |
| `http://` | `http://config.internal/token` | Read from an HTTP endpoint. |
| `https://` | `https://config.example.com/token` | Read from an HTTPS endpoint. |
| `env://` | `env://APP_PASSWORD` | Read a process environment variable. |

`file:///absolute/path` reads an absolute path.
`file://relative/path` reads a path relative to the process working directory;
URI authorities such as `file://host/path` are not supported.

## Set a timeout for network references

Set a timeout when you use `http://` or `https://` references:

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithTimeout(5 * time.Second).
    Build()
```

## Keep source fields before dependent fields

Fuda processes struct fields in declaration order.
Declare fields that provide URI template values before the field that uses `ref` or `refFrom`.

## Next step

[Compose a DSN from resolved values](dsn.md).
