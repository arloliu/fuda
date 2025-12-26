package watcher

import (
	"io"
	"os"
	"time"

	"github.com/arloliu/fuda"
	"github.com/go-playground/validator/v10"
)

// Builder provides a fluent API for constructing a Watcher.
type Builder struct {
	config watcherConfig
	source []byte
	path   string
	err    error
}

// FromFile sets the configuration file to watch.
// The file is monitored for changes using fsnotify.
func (b *Builder) FromFile(path string) *Builder {
	if b.err != nil {
		return b
	}

	data, err := os.ReadFile(path)
	if err != nil {
		b.err = err
		return b
	}

	b.source = data
	b.path = path

	return b
}

// FromReader reads initial configuration from an io.Reader.
// Note: Reader-based sources cannot be watched for changes;
// only polling of ref/refFrom sources will work.
func (b *Builder) FromReader(r io.Reader) *Builder {
	if b.err != nil {
		return b
	}

	data, err := io.ReadAll(r)
	if err != nil {
		b.err = err
		return b
	}

	b.source = data

	return b
}

// FromBytes sets the initial configuration from a byte slice.
// Note: Byte sources cannot be watched for changes;
// only polling of ref/refFrom sources will work.
func (b *Builder) FromBytes(data []byte) *Builder {
	b.source = data
	return b
}

// WithRefResolver sets the reference resolver for ref/refFrom tags.
// The resolver is also used for watching remote secrets if it implements
// the WatchableResolver interface.
func (b *Builder) WithRefResolver(r fuda.RefResolver) *Builder {
	b.config.refResolver = r
	return b
}

// WithEnvPrefix sets a prefix for environment variable lookups.
func (b *Builder) WithEnvPrefix(prefix string) *Builder {
	b.config.envPrefix = prefix
	return b
}

// WithValidator sets a custom validator instance.
func (b *Builder) WithValidator(v *validator.Validate) *Builder {
	b.config.validator = v
	return b
}

// WithWatchInterval sets the polling interval for remote secrets.
// This is the interval at which the watcher checks for changes in
// secrets resolved via ref/refFrom tags (e.g., Vault secrets).
//
// Default is 30 seconds.
func (b *Builder) WithWatchInterval(interval time.Duration) *Builder {
	b.config.watchInterval = interval
	return b
}

// WithDebounceInterval sets the debounce interval for file changes.
// Multiple rapid changes are coalesced into a single reload.
//
// Default is 100 milliseconds.
func (b *Builder) WithDebounceInterval(interval time.Duration) *Builder {
	b.config.debounceInterval = interval
	return b
}

// WithAutoRenewLease enables automatic lease renewal for Vault dynamic secrets.
// When enabled, the watcher will attempt to renew leases before they expire,
// rather than waiting for expiry and re-fetching.
//
// Default is false (no auto-renewal).
func (b *Builder) WithAutoRenewLease() *Builder {
	b.config.autoRenewLease = true
	return b
}

// Build creates the Watcher with the configured options.
func (b *Builder) Build() (*Watcher, error) {
	if b.err != nil {
		return nil, b.err
	}

	// Create the underlying fuda.Loader
	loaderBuilder := fuda.New()

	if b.path != "" {
		loaderBuilder = loaderBuilder.FromFile(b.path)
	} else if len(b.source) > 0 {
		loaderBuilder = loaderBuilder.FromBytes(b.source)
	}

	if b.config.envPrefix != "" {
		loaderBuilder = loaderBuilder.WithEnvPrefix(b.config.envPrefix)
	}

	if b.config.refResolver != nil {
		loaderBuilder = loaderBuilder.WithRefResolver(b.config.refResolver)
	}

	if b.config.validator != nil {
		if v, ok := b.config.validator.(*validator.Validate); ok {
			loaderBuilder = loaderBuilder.WithValidator(v)
		}
	}

	loader, err := loaderBuilder.Build()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		loader:        loader,
		config:        b.config,
		configPath:    b.path,
		configContent: b.source,
	}, nil
}
