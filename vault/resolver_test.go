package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockVaultServer creates a test server that simulates Vault API responses.
func mockVaultServer(t *testing.T, responses map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for response based on path
		if resp, ok := responses[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Errorf("failed to encode response: %v", err)
			}
			return
		}
		// Return a proper 404 for missing secrets
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestNewResolver(t *testing.T) {
	t.Run("requires address", func(t *testing.T) {
		_, err := NewResolver()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "address is required")
	})

	t.Run("creates resolver with address and token", func(t *testing.T) {
		resolver, err := NewResolver(
			WithAddress("https://vault.example.com:8200"),
			WithToken("test-token"),
		)
		require.NoError(t, err)
		assert.NotNil(t, resolver)
		assert.NotNil(t, resolver.Client())
	})

	t.Run("creates resolver with namespace", func(t *testing.T) {
		resolver, err := NewResolver(
			WithAddress("https://vault.example.com:8200"),
			WithToken("test-token"),
			WithNamespace("my-namespace"),
		)
		require.NoError(t, err)
		assert.Equal(t, "my-namespace", resolver.namespace)
	})
}

func TestResolver_Resolve(t *testing.T) {
	t.Run("resolves KV v2 secret", func(t *testing.T) {
		// KV v2 returns data nested under "data" key
		server := mockVaultServer(t, map[string]any{
			"/v1/secret/data/myapp": map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"password": "super-secret",
						"username": "admin",
					},
				},
			},
		})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithToken("test-token"),
		)
		require.NoError(t, err)

		data, err := resolver.Resolve(context.Background(), "vault:///secret/data/myapp#password")
		require.NoError(t, err)
		assert.Equal(t, "super-secret", string(data))
	})

	t.Run("resolves KV v1 secret", func(t *testing.T) {
		// KV v1 returns data at top level
		server := mockVaultServer(t, map[string]any{
			"/v1/kv/myapp": map[string]any{
				"data": map[string]any{
					"api_key": "key-12345",
				},
			},
		})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithToken("test-token"),
		)
		require.NoError(t, err)

		data, err := resolver.Resolve(context.Background(), "vault:///kv/myapp#api_key")
		require.NoError(t, err)
		assert.Equal(t, "key-12345", string(data))
	})

	t.Run("returns error for missing secret", func(t *testing.T) {
		server := mockVaultServer(t, map[string]any{})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithToken("test-token"),
		)
		require.NoError(t, err)

		_, err = resolver.Resolve(context.Background(), "vault:///secret/data/nonexistent#field")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for missing field", func(t *testing.T) {
		server := mockVaultServer(t, map[string]any{
			"/v1/secret/data/myapp": map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"password": "secret",
					},
				},
			},
		})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithToken("test-token"),
		)
		require.NoError(t, err)

		_, err = resolver.Resolve(context.Background(), "vault:///secret/data/myapp#nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for invalid URI", func(t *testing.T) {
		resolver, err := NewResolver(
			WithAddress("https://vault.example.com:8200"),
			WithToken("test-token"),
		)
		require.NoError(t, err)

		_, err = resolver.Resolve(context.Background(), "vault:///")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing")
	})

	t.Run("returns error for wrong scheme", func(t *testing.T) {
		resolver, err := NewResolver(
			WithAddress("https://vault.example.com:8200"),
			WithToken("test-token"),
		)
		require.NoError(t, err)

		_, err = resolver.Resolve(context.Background(), "http://example.com/secret#field")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported scheme")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		resolver, err := NewResolver(
			WithAddress("https://vault.example.com:8200"),
			WithToken("test-token"),
		)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err = resolver.Resolve(ctx, "vault:///secret/data/myapp#password")
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestResolver_AuthMethods(t *testing.T) {
	t.Run("kubernetes auth", func(t *testing.T) {
		// Create a temp file to simulate the JWT
		tmpFile, err := os.CreateTemp("", "k8s-token-*")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("test-jwt-token")
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		// Mock server that handles both auth and secret read
		server := mockVaultServer(t, map[string]any{
			"/v1/auth/kubernetes/login": map[string]any{
				"auth": map[string]any{
					"client_token": "authenticated-token",
				},
			},
			"/v1/secret/data/myapp": map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"password": "k8s-secret",
					},
				},
			},
		})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithKubernetesAuth("my-role", tmpFile.Name()),
		)
		require.NoError(t, err)

		data, err := resolver.Resolve(context.Background(), "vault:///secret/data/myapp#password")
		require.NoError(t, err)
		assert.Equal(t, "k8s-secret", string(data))
	})

	t.Run("approle auth", func(t *testing.T) {
		server := mockVaultServer(t, map[string]any{
			"/v1/auth/approle/login": map[string]any{
				"auth": map[string]any{
					"client_token": "approle-token",
				},
			},
			"/v1/secret/data/myapp": map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"password": "approle-secret",
					},
				},
			},
		})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithAppRole("role-id", "secret-id"),
		)
		require.NoError(t, err)

		data, err := resolver.Resolve(context.Background(), "vault:///secret/data/myapp#password")
		require.NoError(t, err)
		assert.Equal(t, "approle-secret", string(data))
	})

	t.Run("requires auth method when no token", func(t *testing.T) {
		resolver, err := NewResolver(
			WithAddress("https://vault.example.com:8200"),
		)
		require.NoError(t, err)

		_, err = resolver.Resolve(context.Background(), "vault:///secret/data/myapp#password")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no authentication method")
	})
}

func TestWithOptions(t *testing.T) {
	t.Run("WithKubernetesAuthMount", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "k8s-token-*")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		_, _ = tmpFile.WriteString("test-jwt")
		tmpFile.Close()

		server := mockVaultServer(t, map[string]any{
			"/v1/auth/custom-k8s/login": map[string]any{
				"auth": map[string]any{
					"client_token": "token",
				},
			},
			"/v1/secret/data/test": map[string]any{
				"data": map[string]any{
					"data": map[string]any{"key": "value"},
				},
			},
		})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithKubernetesAuthMount("custom-k8s", "role", tmpFile.Name()),
		)
		require.NoError(t, err)

		data, err := resolver.Resolve(context.Background(), "vault:///secret/data/test#key")
		require.NoError(t, err)
		assert.Equal(t, "value", string(data))
	})

	t.Run("WithAppRoleMount", func(t *testing.T) {
		server := mockVaultServer(t, map[string]any{
			"/v1/auth/custom-approle/login": map[string]any{
				"auth": map[string]any{
					"client_token": "token",
				},
			},
			"/v1/secret/data/test": map[string]any{
				"data": map[string]any{
					"data": map[string]any{"key": "value"},
				},
			},
		})
		defer server.Close()

		resolver, err := NewResolver(
			WithAddress(server.URL),
			WithAppRoleMount("custom-approle", "role-id", "secret-id"),
		)
		require.NoError(t, err)

		data, err := resolver.Resolve(context.Background(), "vault:///secret/data/test#key")
		require.NoError(t, err)
		assert.Equal(t, "value", string(data))
	})
}
