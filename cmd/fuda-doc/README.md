# fuda-doc

> 📜 A documentation generator for fuda configuration structs

**fuda-doc** is a CLI tool that parses Go source files and generates documentation for configuration structs that use fuda's struct tags (`default`, `env`, `ref`, `dsn`, etc.).

## Features

- **Multiple Output Formats**
  - **ASCII** — Terminal-friendly output with ANSI colors and a built-in pager
  - **Markdown** — GitHub-compatible Markdown for documentation sites
  - **YAML** — Default configuration file generation with comments
  - **.env** — Environment variable template file generation

- **Interactive TUI Explorer** — Browse all configuration structs interactively using a tree-based UI with search and filtering

- **Struct Tag Extraction** — Automatically extracts and documents:
  - Default values (`default` tag)
  - Environment variable bindings (`env` tag)
  - External references (`ref`, `refFrom` tags)
  - DSN composition (`dsn` tag)
  - Validation rules (`validate` tag)
  - Nested struct support with full hierarchy

## Installation

```bash
go install github.com/arloliu/fuda/cmd/fuda-doc@latest
```

Or build from source:

```bash
cd cmd/fuda-doc
make build
```

## Usage

### Basic Documentation Generation

```bash
# Generate ASCII documentation (default)
fuda-doc -struct Config -path ./internal/config

# Generate Markdown documentation
fuda-doc -struct Config -path ./internal/config --markdown

# Output to a file
fuda-doc -struct Config -path ./internal/config --markdown -o CONFIG.md
```

### Interactive TUI Mode

```bash
# Explore all structs in a package
fuda-doc -tui -path ./internal/config

# Explore a specific struct
fuda-doc -tui -struct Config -path ./internal/config
```

The TUI provides:
- Tree navigation of nested configuration fields
- Detail panel with field information (type, tags, description)
- YAML preview panel
- Search functionality (`/` to search)
- Tag filtering
- Export options (copy to clipboard, save to file)

### Utility Modes

```bash
# Generate environment variable summary table
fuda-doc --env-summary -path ./internal/config

# Generate .env.example template
fuda-doc --env-file -path ./internal/config

# Generate default YAML configuration with comments
fuda-doc --yaml-default -path ./internal/config
```

## Command Reference

| Flag             | Short | Description                                                   |
|------------------|-------|---------------------------------------------------------------|
| `--struct`       | `-s`  | Struct name to generate docs for (required unless `-tui`)     |
| `--path`         | `-p`  | Directory or file path containing the struct (required)       |
| `--output`       | `-o`  | Output target: file path or "stdout" (default: stdout)        |
| `--markdown`     | `-m`  | Output in Markdown format                                     |
| `--ascii`        | `-a`  | Output in terminal-friendly format with ANSI colors (default) |
| `--no-pager`     |       | Disable built-in pager for ASCII output                       |
| `--color`        | `-c`  | Force ANSI color output (useful with: `\| less -R`)           |
| `--tui`          | `-t`  | Launch interactive TUI explorer                               |
| `--version`      | `-v`  | Print version and exit                                        |
| `--env-summary`  |       | Print a summary table of all env-tagged fields                |
| `--env-file`     |       | Generate a .env.example file from env-tagged fields           |
| `--yaml-default` |       | Generate a default YAML config with comments                  |

## Example Output

Given a configuration struct:

```go
// AppConfig holds application configuration.
type AppConfig struct {
    Host    string        `yaml:"host" default:"localhost" env:"APP_HOST"`
    Port    int           `yaml:"port" default:"8080" env:"APP_PORT"`
    Timeout time.Duration `yaml:"timeout" default:"30s"`
    Debug   bool          `yaml:"debug" default:"false" env:"APP_DEBUG"`
}
```

Running `fuda-doc -struct AppConfig -path .` produces documentation including:

- Usage example with code snippet
- Configuration resolution order
- YAML example with default values
- Field reference table with type, default, env var, and description

## TUI Keyboard Shortcuts

| Key       | Action                  |
|-----------|-------------------------|
| `↑` / `k` | Move up                 |
| `↓` / `j` | Move down               |
| `←` / `h` | Collapse / go to parent |
| `→` / `l` | Expand / enter          |
| `Tab`     | Switch panels           |
| `/`       | Start search            |
| `f`       | Filter by tag           |
| `e`       | Export menu             |
| `?`       | Toggle help             |
| `q`       | Quit                    |

## License

MIT License — see [LICENSE](../../LICENSE) for details.
