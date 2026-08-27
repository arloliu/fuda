# Reload configuration with Watcher

Use `fuda/watcher` when a long-running application should reload configuration after a file or remote reference changes.

```go
configWatcher, err := watcher.New().
    FromFile("config.yaml").
    Build()
if err != nil {
    return err
}
defer configWatcher.Stop()

var cfg Config
updates, err := configWatcher.Watch(&cfg)
if err != nil {
    return err
}
```

`Watch(&cfg)` loads the initial value into `cfg`.
It returns a channel for later changes.

```go
go func() {
    for update := range updates {
        next := update.(*Config)
        application.ApplyConfig(next)
    }
}()
```

Move the new snapshot into your application's own synchronized handoff.
Do not let unrelated goroutines mutate the same config pointer.

## What Watcher observes

For a file source, Watcher reacts to file write and create events.
It debounces rapid changes before it reloads the file.
For remote references, it polls and sends an update only when the loaded configuration changed.

Set the polling interval when your resolver reads remote values:

```go
configWatcher, err := watcher.New().
    FromFile("config.yaml").
    WithWatchInterval(30 * time.Second).
    Build()
```

Call `Stop()` during shutdown.
Watcher closes the updates channel after its watch loop exits.

## Next step

[Load secrets from HashiCorp Vault](vault.md).
