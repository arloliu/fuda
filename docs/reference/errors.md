# Errors

Fuda returns ordinary Go errors from builder operations and configuration loading.
Check errors during startup and include enough context for an operator to fix the input.

```go
var cfg Config
if err := loader.Load(&cfg); err != nil {
    return fmt.Errorf("load configuration: %w", err)
}
```

Use `errors.As` when you need structured details.

```go
var fieldErr *fuda.FieldError
if errors.As(err, &fieldErr) {
    log.Printf("field=%s tag=%s", fieldErr.Path, fieldErr.Tag)
}
```

## Public error types

| Type | Use |
| --- | --- |
| `fuda.FieldError` | Identifies a field and tag that failed conversion or processing. |
| `fuda.LoadError` | Holds multiple field errors for a load operation. |
| `fuda.ValidationError` | Wraps validation errors after loading. |

Builder source errors include missing files and read failures.
Loader errors can also report YAML decoding, invalid tag values, reference resolution, and validation failures.

Do not parse error strings for application logic.
Use `errors.As` and the exposed fields instead.

## Next step

[Troubleshoot a common failure](troubleshooting.md).
