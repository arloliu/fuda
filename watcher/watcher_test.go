package watcher

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Host    string `yaml:"host" default:"localhost"`
	Port    int    `yaml:"port" default:"8080"`
	Timeout string `yaml:"timeout" default:"30s"`
}

func TestWatcher_New(t *testing.T) {
	t.Run("creates watcher from file", func(t *testing.T) {
		// Create temp config file
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("host: example.com\nport: 9090\n")
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		w, err := New().
			FromFile(tmpFile.Name()).
			Build()
		require.NoError(t, err)
		require.NotNil(t, w)
		defer w.Stop()
	})

	t.Run("creates watcher from bytes", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: example.com\n")).
			Build()
		require.NoError(t, err)
		require.NotNil(t, w)
		defer w.Stop()
	})

	t.Run("fails for nonexistent file", func(t *testing.T) {
		_, err := New().
			FromFile("/nonexistent/config.yaml").
			Build()
		require.Error(t, err)
	})
}

func TestWatcher_Watch(t *testing.T) {
	t.Run("loads initial config", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("host: initial.com\nport: 1234\n")
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		w, err := New().
			FromFile(tmpFile.Name()).
			Build()
		require.NoError(t, err)
		defer w.Stop()

		var cfg testConfig
		updates, err := w.Watch(&cfg)
		require.NoError(t, err)
		require.NotNil(t, updates)

		assert.Equal(t, "initial.com", cfg.Host)
		assert.Equal(t, 1234, cfg.Port)
	})

	t.Run("detects file changes", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("host: initial.com\nport: 1234\n")
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		w, err := New().
			FromFile(tmpFile.Name()).
			WithWatchInterval(50 * time.Millisecond).
			WithDebounceInterval(10 * time.Millisecond).
			Build()
		require.NoError(t, err)
		defer w.Stop()

		var cfg testConfig
		updates, err := w.Watch(&cfg)
		require.NoError(t, err)

		// Initial values
		assert.Equal(t, "initial.com", cfg.Host)

		// Give fsnotify time to set up the watch
		time.Sleep(50 * time.Millisecond)

		// Modify the file
		err = os.WriteFile(tmpFile.Name(), []byte("host: updated.com\nport: 5678\n"), 0o644)
		require.NoError(t, err)

		// Wait for update
		select {
		case newCfg := <-updates:
			updatedCfg, ok := newCfg.(*testConfig)
			require.True(t, ok, "expected *testConfig")
			assert.Equal(t, "updated.com", updatedCfg.Host)
			assert.Equal(t, 5678, updatedCfg.Port)
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for config update")
		}
	})

	t.Run("prevents double watch", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: test\n")).
			Build()
		require.NoError(t, err)
		defer w.Stop()

		var cfg testConfig
		_, err = w.Watch(&cfg)
		require.NoError(t, err)

		// Try to watch again
		var cfg2 testConfig
		_, err = w.Watch(&cfg2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already running")
	})
}

func TestWatcher_Stop(t *testing.T) {
	t.Run("stops gracefully", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: test\n")).
			Build()
		require.NoError(t, err)

		var cfg testConfig
		updates, err := w.Watch(&cfg)
		require.NoError(t, err)

		// Stop the watcher
		w.Stop()

		// Channel should be closed
		select {
		case _, ok := <-updates:
			assert.False(t, ok, "channel should be closed")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for channel close")
		}
	})

	t.Run("stop is idempotent", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: test\n")).
			Build()
		require.NoError(t, err)

		var cfg testConfig
		_, err = w.Watch(&cfg)
		require.NoError(t, err)

		// Multiple stops should not panic
		w.Stop()
		w.Stop()
		w.Stop()
	})
}

func TestWatcher_Concurrent(t *testing.T) {
	t.Run("handles concurrent access", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("host: test.com\nport: 8080\n")
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		w, err := New().
			FromFile(tmpFile.Name()).
			WithWatchInterval(100 * time.Millisecond).
			WithDebounceInterval(20 * time.Millisecond).
			Build()
		require.NoError(t, err)

		var cfg testConfig
		updates, err := w.Watch(&cfg)
		require.NoError(t, err)

		// Track updates received
		var updateCount int64

		// Consumer goroutine
		done := make(chan struct{})
		go func() {
			for range updates {
				atomic.AddInt64(&updateCount, 1)
			}
			close(done)
		}()

		// Give fsnotify time to set up
		time.Sleep(100 * time.Millisecond)

		// Make a significant change to trigger an update
		err = os.WriteFile(tmpFile.Name(), []byte("host: updated.com\nport: 9999\n"), 0o644)
		require.NoError(t, err)

		// Give time for update to propagate
		time.Sleep(500 * time.Millisecond)

		w.Stop()
		<-done

		// Should have received at least one update
		assert.GreaterOrEqual(t, atomic.LoadInt64(&updateCount), int64(1))
	})
}

func TestBuilder_Options(t *testing.T) {
	t.Run("WithWatchInterval", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: test\n")).
			WithWatchInterval(5 * time.Minute).
			Build()
		require.NoError(t, err)
		defer w.Stop()

		assert.Equal(t, 5*time.Minute, w.config.watchInterval)
	})

	t.Run("WithDebounceInterval", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: test\n")).
			WithDebounceInterval(500 * time.Millisecond).
			Build()
		require.NoError(t, err)
		defer w.Stop()

		assert.Equal(t, 500*time.Millisecond, w.config.debounceInterval)
	})

	t.Run("WithEnvPrefix", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: test\n")).
			WithEnvPrefix("APP_").
			Build()
		require.NoError(t, err)
		defer w.Stop()

		assert.Equal(t, "APP_", w.config.envPrefix)
	})

	t.Run("WithAutoRenewLease", func(t *testing.T) {
		w, err := New().
			FromBytes([]byte("host: test\n")).
			WithAutoRenewLease().
			Build()
		require.NoError(t, err)
		defer w.Stop()

		assert.True(t, w.config.autoRenewLease)
	})
}
