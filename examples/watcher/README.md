# Hot-Reload Watcher Example

Demonstrates using the watcher package for automatic configuration reloading.

## Features

- File system watching with fsnotify
- Debouncing rapid changes
- Thread-safe config access with atomic.Pointer
- Graceful shutdown

## Run

```bash
go run main.go
```

Then edit the generated `config.yaml` in another terminal to see hot-reload in action.

## Use Case

Perfect for long-running services that need to reload configuration without restart:

```go
// Access current config from any goroutine
cfg := globalConfig.Load()
fmt.Printf("Max connections: %d\n", cfg.MaxConns)
```
