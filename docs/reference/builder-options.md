# Builder options

Start every configurable load with `fuda.New()`.

| Method | Purpose |
| --- | --- |
| `FromFile(path)` | Reads YAML or JSON from a file. |
| `FromReader(reader)` | Reads source bytes from an `io.Reader`. |
| `FromBytes(data)` | Uses source bytes directly. |
| `WithEnvPrefix(prefix)` | Prepends a prefix to `env` tags. |
| `WithValidator(validator)` | Replaces the default validator. |
| `WithRefResolver(resolver)` | Replaces the default reference resolver. |
| `WithFilesystem(fs)` | Uses an afero filesystem. |
| `WithTimeout(timeout)` | Limits reference resolution time. |
| `WithOverrides(values)` | Changes rendered source before decoding. |
| `WithSizePreprocess(enabled)` | Enables or disables numeric byte-size conversion. |
| `WithDurationPreprocess(enabled)` | Enables or disables duration-day conversion. |
| `Apply(func)` | Applies reusable builder configuration. |
| `WithTemplate(data, options...)` | Renders source as a Go template. |
| `WithDotEnv(file, options...)` | Loads one dotenv file. |
| `WithDotEnvFiles(files, options...)` | Loads dotenv overlays. |
| `WithDotEnvSearch(name, paths, options...)` | Searches for one dotenv file. |
| `Build()` | Returns the configured loader. |

```go
loader, err := fuda.New().
    FromFile("config.yaml").
    WithEnvPrefix("APP_").
    WithTimeout(5 * time.Second).
    Build()
```

Call `Load(&cfg)` on the returned loader.

## Next step

[Review the loading behaviour](loading-behaviour.md).
