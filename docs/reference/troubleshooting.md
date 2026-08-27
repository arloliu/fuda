# Troubleshooting

## My default did not apply

Check whether YAML supplied an explicit value.
Fuda keeps `0`, `false`, and `""` because those values express an operator choice.
Use YAML `null` or remove the key when you want the default.

## My environment variable did not work

Check the exact `env` tag name.
If you use `WithEnvPrefix("APP_")` with `env:"PORT"`, set `APP_PORT`.
Confirm the value can convert to the Go field type.

## My reference timed out or failed

Check the URI scheme and the process permissions for `file://` paths.
Set `WithTimeout` for HTTP references.
Use a custom resolver for a scheme that the default resolver does not support.

## My DSN has an empty segment

Put dependency fields before the DSN field.
Add `dsnStrict:"true"` when a missing value must stop startup.

## My configuration failed validation

Read the wrapped `fuda.ValidationError`.
The message identifies the invalid field and the violated rule.

## Next step

[Return to the Builder options](builder-options.md).
