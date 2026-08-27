# Generate configuration documentation

`fuda-doc` reads Go source and produces documentation for tagged configuration structs.

## Install it

```bash
go install github.com/arloliu/fuda/cmd/fuda-doc@latest
```

## Generate Markdown first

Generate Markdown for one struct and write it to a file:

```bash
fuda-doc --struct Config --path ./internal/config --markdown --output CONFIG.md
```

Use `--struct` or `-s` for the struct name.
Use `--path` or `-p` for the file or directory that contains it.
`--output` or `-o` defaults to `stdout`.

## Other output modes

Use `--ascii` or `-a` for terminal output.
Use `--no-pager` to disable the pager for ASCII output.
Use `--color` or `-c` to force ANSI colors.

Use `--env-summary`, `--env-file`, or `--yaml-default` with `--path` to generate supporting configuration artifacts.
Use `--tui` or `-t` to browse structs interactively.
Use `--version` or `-v` to print the installed version.

## Next step

[Review the generated tags](../reference/tags.md).
