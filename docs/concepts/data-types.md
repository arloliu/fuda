# Data types

Fuda converts common scalar values and adds helpers for readable durations and byte sizes.

```go
type Config struct {
    Timeout     time.Duration `yaml:"timeout"`
    Retention   fuda.Duration `yaml:"retention"`
    UploadLimit int64         `yaml:"upload_limit"`
    CacheSize   fuda.ByteSize `yaml:"cache_size"`
}
```

```yaml
timeout: 30s
retention: 7d
upload_limit: 64KiB
cache_size: 10MiB
```

## Durations

Use `time.Duration` for standard Go duration values such as `30s`, `5m`, and `1h30m`.
Fuda also accepts a `d` suffix for whole or fractional days when it loads a `time.Duration` field.

Use `fuda.Duration` when you want YAML or JSON output to keep a readable duration string.
Call `Duration()` when you need the underlying `time.Duration`.

## Byte sizes

Fuda converts size strings such as `64KiB` and `10MiB` to byte counts for numeric fields such as `int64`.
Use `fuda.ByteSize` when you also want readable YAML and JSON serialization.
Call `Int64()` to get the byte count.

The size and duration preprocessing options are enabled by default.
Turn them off with `WithSizePreprocess(false)` or `WithDurationPreprocess(false)` when your input format needs different handling.

## Next step

[Return to the precedence rule](../getting-started/choosing-a-value.md) before you add more configuration sources.
