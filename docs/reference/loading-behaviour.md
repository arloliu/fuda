# Loading behaviour

Fuda applies configuration in a fixed sequence.
Read this page when you combine several features.

1. Load configured dotenv files.
2. Render the source with `WithTemplate`, if configured.
3. Apply `WithOverrides` to the rendered source.
4. Decode the resulting YAML or JSON into the target struct.
5. Process each field's `env`, `refFrom` or `ref`, `default`, and `dsn` tags.
6. Call `SetDefaults()` on matching structs, then validate the final target.

## Field precedence

For an ordinary field, an environment value wins over a configuration-file value.
The file value wins over a default.
Fuda preserves explicit zero values such as `0`, `false`, and `""` from the file.
YAML `null` behaves as absent for default handling.

`refFrom` runs before its fallback `ref` tag when a field is still zero.
A reference can provide a value before Fuda considers a default.

`dsn` runs after the other field tags.
Declare values a DSN needs before the DSN field.

## Setter and validation

Fuda processes nested fields before it calls `SetDefaults()` on the containing struct.
Validation runs last, so it observes file values, environment overrides, references, defaults, DSNs, and dynamic defaults.

## Next step

[Handle loading errors](errors.md).
