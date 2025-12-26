# Configuration Watcher

The `fuda/watcher` package provides hot-reload configuration watching, automatically detecting changes in config files and secrets.

## Quick Start

```go
package main

import (
    "log"
    "sync/atomic"
    "time"
    "unsafe"

    "github.com/arloliu/fuda/watcher"
)

type Config struct {
    Host    string `yaml:"host"`
    Port    int    `yaml:"port"`
    Timeout string `yaml:"timeout"`
}

var globalConfig atomic.Pointer[Config]

func main() {
    w, err := watcher.New().
        FromFile("config.yaml").
        WithWatchInterval(30 * time.Second).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    defer w.Stop()

    var cfg Config
    updates, err := w.Watch(&cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Store initial config
    globalConfig.Store(&cfg)

    // Handle updates in a goroutine
    go func() {
        for newCfg := range updates {
            log.Println("Config updated!")
            globalConfig.Store(newCfg.(*Config))
        }
    }()

    // Use config
    config := globalConfig.Load()
    log.Printf("Server: %s:%d\n", config.Host, config.Port)

    // Keep running...
    select {}
}
```

## Watch Mechanisms

The watcher uses two mechanisms for detecting changes:

| Mechanism | Source | How It Works |
|-----------|--------|--------------|
| **fsnotify** | Config files, local secrets | Real-time file system events |
| **Polling** | Vault, HTTP refs | Periodic checks at `WatchInterval` |

## Builder Options

```go
watcher.New().
    FromFile("config.yaml").              // Watch this file
    WithRefResolver(vaultResolver).        // For vault:// refs
    WithEnvPrefix("APP_").                 // Environment prefix
    WithWatchInterval(30 * time.Second).   // Poll interval for remote refs
    WithDebounceInterval(100 * time.Millisecond). // Coalesce rapid changes
    WithAutoRenewLease().                  // Auto-renew Vault leases
    Build()
```

| Option | Default | Description |
|--------|---------|-------------|
| `WithWatchInterval` | 30s | Polling interval for remote secrets |
| `WithDebounceInterval` | 100ms | Coalesce multiple rapid file changes |
| `WithAutoRenewLease` | false | Auto-renew Vault dynamic secret leases |

## Thread-Safe Config Access

### Using `atomic.Pointer` (Go 1.19+)

```go
var globalConfig atomic.Pointer[Config]

// Store
globalConfig.Store(newConfig)

// Load (safe from any goroutine)
cfg := globalConfig.Load()
```

### Using `sync.RWMutex`

```go
var (
    config   Config
    configMu sync.RWMutex
)

// Update (in watcher goroutine)
configMu.Lock()
config = *newCfg.(*Config)
configMu.Unlock()

// Read (from any goroutine)
configMu.RLock()
host := config.Host
configMu.RUnlock()
```

## Graceful Shutdown

```go
// In your shutdown handler
w.Stop()

// Stop() blocks until:
// - The watch loop terminates
// - The updates channel is closed
// - fsnotify resources are released
```

## Complete Example with Vault

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "sync/atomic"
    "syscall"
    "time"

    "github.com/arloliu/fuda/vault"
    "github.com/arloliu/fuda/watcher"
)

type Config struct {
    Database struct {
        Host     string `yaml:"host" default:"localhost"`
        Password string `ref:"vault:///secret/data/db#password"`
    } `yaml:"database"`
}

var globalConfig atomic.Pointer[Config]

func main() {
    // Create Vault resolver
    vaultResolver, err := vault.NewResolver(
        vault.WithAddress(os.Getenv("VAULT_ADDR")),
        vault.WithKubernetesAuth("my-role", "/var/run/secrets/kubernetes.io/serviceaccount/token"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create watcher
    w, err := watcher.New().
        FromFile("/etc/config/app.yaml").
        WithRefResolver(vaultResolver).
        WithWatchInterval(5 * time.Minute).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    var cfg Config
    updates, err := w.Watch(&cfg)
    if err != nil {
        log.Fatal(err)
    }
    globalConfig.Store(&cfg)

    // Handle updates
    go func() {
        for newCfg := range updates {
            log.Println("Configuration reloaded")
            globalConfig.Store(newCfg.(*Config))
            // Notify application components of config change
        }
    }()

    // Graceful shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    log.Println("Shutting down...")
    w.Stop()
}
```

## Error Handling

The watcher silently continues watching if a reload fails (e.g., invalid YAML, network error). This ensures your application keeps running with the last known good configuration.

To debug reload issues, check:
- File permissions
- YAML/JSON syntax
- Vault connectivity and authentication
- Environment variable values
