// Package watcher provides hot-reload configuration watching for fuda.
//
// This package enables automatic configuration reloading when source files
// or secrets change. It supports file system watching via fsnotify and
// periodic polling for remote secrets (e.g., Vault).
//
// Basic usage:
//
//	watcher, err := watcher.New().
//	    FromFile("config.yaml").
//	    WithWatchInterval(30 * time.Second).
//	    Build()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer watcher.Stop()
//
//	var cfg Config
//	updates, err := watcher.Watch(&cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Handle updates in a goroutine
//	go func() {
//	    for newCfg := range updates {
//	        app.UpdateConfig(newCfg.(*Config))
//	    }
//	}()
//
// # Watch Mechanisms
//
// The watcher uses two mechanisms for detecting changes:
//
// 1. File system watching (fsnotify) - for config files and local secrets
// 2. Periodic polling - for remote secrets (Vault, HTTP endpoints)
//
// # Thread Safety
//
// The Watcher is safe for concurrent use. The updates channel should be
// consumed by a single goroutine to avoid race conditions when updating
// application state.
package watcher

import (
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/arloliu/fuda"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors configuration sources and emits updates when changes occur.
type Watcher struct {
	loader        *fuda.Loader
	config        watcherConfig
	fsWatcher     *fsnotify.Watcher
	stopChan      chan struct{}
	doneChan      chan struct{}
	updatesChan   chan any
	mu            sync.Mutex
	running       bool
	watchedFiles  []string
	lastConfig    any
	configPath    string
	configContent []byte
}

// watcherConfig holds internal configuration for the watcher.
type watcherConfig struct {
	watchInterval    time.Duration
	refResolver      fuda.RefResolver
	envPrefix        string
	autoRenewLease   bool
	debounceInterval time.Duration
	validator        any // *validator.Validate
}

// defaultWatchInterval is the default polling interval for remote secrets.
const defaultWatchInterval = 30 * time.Second

// defaultDebounceInterval prevents rapid successive reloads.
const defaultDebounceInterval = 100 * time.Millisecond

// New creates a new watcher Builder.
func New() *Builder {
	return &Builder{
		config: watcherConfig{
			watchInterval:    defaultWatchInterval,
			debounceInterval: defaultDebounceInterval,
		},
	}
}

// Watch starts watching for configuration changes.
// It returns a channel that receives new configuration values when changes occur.
// The initial configuration is loaded and returned before the channel starts emitting updates.
//
// The returned channel is closed when Stop() is called.
// The caller should consume the channel in a goroutine:
//
//	updates, err := watcher.Watch(&cfg)
//	go func() {
//	    for newCfg := range updates {
//	        app.UpdateConfig(newCfg.(*Config))
//	    }
//	}()
func (w *Watcher) Watch(target any) (<-chan any, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return nil, &WatcherError{Message: "watcher is already running"}
	}

	// Perform initial load
	if err := w.loader.Load(target); err != nil {
		return nil, err
	}

	// Store a copy of the initial config for change detection
	w.lastConfig = w.deepCopy(target)

	// Start watching
	w.running = true
	w.updatesChan = make(chan any, 1)
	w.stopChan = make(chan struct{})
	w.doneChan = make(chan struct{})

	go w.watchLoop(target)

	return w.updatesChan, nil
}

// Stop gracefully stops the watcher.
// It closes the updates channel and releases resources.
func (w *Watcher) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	w.mu.Unlock()

	close(w.stopChan)
	<-w.doneChan // Wait for watchLoop to finish

	if w.fsWatcher != nil {
		_ = w.fsWatcher.Close()
	}
}

// watchLoop is the main watch loop that monitors for changes.
func (w *Watcher) watchLoop(target any) {
	defer close(w.doneChan)
	defer close(w.updatesChan)

	// Setup file watcher if we have a config file
	var fsChan <-chan fsnotify.Event
	if w.configPath != "" {
		var err error
		w.fsWatcher, err = fsnotify.NewWatcher()
		if err == nil {
			_ = w.fsWatcher.Add(w.configPath)
			fsChan = w.fsWatcher.Events
			w.watchedFiles = append(w.watchedFiles, w.configPath)
		}
	}

	// Setup polling timer for remote secrets
	pollTicker := time.NewTicker(w.config.watchInterval)
	defer pollTicker.Stop()

	// Debounce timer to prevent rapid successive reloads
	var debounceTimer *time.Timer
	var debounceChan <-chan time.Time

	reload := func() {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.NewTimer(w.config.debounceInterval)
		debounceChan = debounceTimer.C
	}

	for {
		select {
		case <-w.stopChan:
			return

		case event, ok := <-fsChan:
			if !ok {
				fsChan = nil
				continue
			}
			// Only react to write and create events
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				reload()
			}

		case <-pollTicker.C:
			// Poll remote secrets
			reload()

		case <-debounceChan:
			debounceChan = nil
			if changed := w.reloadIfChanged(target); changed {
				// Create a copy and send to updates channel
				newConfig := w.deepCopy(target)
				select {
				case w.updatesChan <- newConfig:
				case <-w.stopChan:
					return
				}
			}
		}
	}
}

// reloadIfChanged reloads configuration and returns true if it changed.
func (w *Watcher) reloadIfChanged(target any) bool {
	// For file-based config, check if content changed
	if w.configPath != "" {
		content, err := os.ReadFile(w.configPath)
		if err != nil {
			return false
		}
		// Quick check: if content is identical, skip full reload
		if string(content) == string(w.configContent) {
			return false
		}
		w.configContent = content
	}

	// Create a new target of the same type
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return false
	}
	newTarget := reflect.New(targetType.Elem()).Interface()

	// Create a fresh loader with updated content for file-based config
	var loadErr error
	if w.configPath != "" && len(w.configContent) > 0 {
		// Create a new loader with the updated content
		builder := fuda.New().FromBytes(w.configContent)
		if w.config.envPrefix != "" {
			builder = builder.WithEnvPrefix(w.config.envPrefix)
		}
		if w.config.refResolver != nil {
			builder = builder.WithRefResolver(w.config.refResolver)
		}
		freshLoader, err := builder.Build()
		if err != nil {
			return false
		}
		loadErr = freshLoader.Load(newTarget)
	} else {
		loadErr = w.loader.Load(newTarget)
	}

	if loadErr != nil {
		// Log error but don't stop watching
		return false
	}

	// Compare with last config
	if w.configEquals(newTarget, w.lastConfig) {
		return false
	}

	// Update target in place
	reflect.ValueOf(target).Elem().Set(reflect.ValueOf(newTarget).Elem())
	w.lastConfig = w.deepCopy(target)

	return true
}

// deepCopy creates a deep copy of the config value.
func (w *Watcher) deepCopy(v any) any {
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Create a new instance and copy
	newVal := reflect.New(val.Type())
	newVal.Elem().Set(val)

	return newVal.Interface()
}

// configEquals compares two config values for equality.
func (w *Watcher) configEquals(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

// WatcherError represents a watcher-specific error.
type WatcherError struct {
	Message string
	Err     error
}

func (e *WatcherError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *WatcherError) Unwrap() error {
	return e.Err
}
