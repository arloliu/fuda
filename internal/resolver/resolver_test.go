package resolver_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arloliu/fuda/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileResolver(t *testing.T) {
	r := resolver.NewFileResolver(nil)
	ctx := context.Background()

	t.Run("valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(file, []byte("content"), 0o600)
		require.NoError(t, err)

		content, err := r.Resolve(ctx, "file://"+file)
		require.NoError(t, err)
		assert.Equal(t, []byte("content"), content)
	})

	t.Run("invalid scheme", func(t *testing.T) {
		_, err := r.Resolve(ctx, "http://example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported scheme")
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := r.Resolve(ctx, "file:///non/existent")
		assert.Error(t, err)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(file, []byte("content"), 0o600)
		require.NoError(t, err)

		_, err = r.Resolve(ctx, "file://"+file)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestHTTPResolver(t *testing.T) {
	r := resolver.NewHTTPResolver()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/timeout" {
			time.Sleep(200 * time.Millisecond)
		}
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
		_, _ = fmt.Fprint(w, "response")
	}))
	defer ts.Close()

	t.Run("valid url", func(t *testing.T) {
		content, err := r.Resolve(ctx, ts.URL)
		require.NoError(t, err)
		assert.Equal(t, []byte("response"), content)
	})

	t.Run("http error", func(t *testing.T) {
		_, err := r.Resolve(ctx, ts.URL+"/error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status: 500")
	})

	t.Run("invalid scheme", func(t *testing.T) {
		_, err := r.Resolve(ctx, "ftp://example.com")
		assert.Error(t, err)
	})

	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err := r.Resolve(ctx, ts.URL+"/timeout")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

func TestCompositeResolver(t *testing.T) {
	r := resolver.New(nil)
	ctx := context.Background()

	t.Run("default schemes", func(_ *testing.T) {
		// Just check that it accepts the schemes, integration test covers actual functionality
		// We can mock sub-resolvers if we want to test delegation purely.
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		_, err := r.Resolve(ctx, "ftp://example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported scheme")
	})

	t.Run("malformed uri", func(t *testing.T) {
		_, err := r.Resolve(ctx, "invalid-uri")
		assert.Error(t, err)
	})
}
