package vault

import (
	"context"
	"errors"
	"fmt"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

// Option configures a Vault resolver.
type Option func(*resolverConfig)

// WithAddress sets the Vault server address.
// This is required for creating a resolver.
//
// Example:
//
//	vault.WithAddress("https://vault.example.com:8200")
func WithAddress(addr string) Option {
	return func(c *resolverConfig) {
		c.address = addr
	}
}

// WithToken sets a static token for authentication.
// This is the simplest authentication method, suitable for development
// or when tokens are injected via environment variables.
//
// Example:
//
//	vault.WithToken(os.Getenv("VAULT_TOKEN"))
func WithToken(token string) Option {
	return func(c *resolverConfig) {
		c.token = token
	}
}

// WithNamespace sets the Vault namespace (Enterprise feature).
// Namespaces provide tenant isolation in Vault Enterprise.
//
// Example:
//
//	vault.WithNamespace("my-team")
func WithNamespace(ns string) Option {
	return func(c *resolverConfig) {
		c.namespace = ns
	}
}

// WithTLSConfig sets custom TLS configuration for the Vault client.
//
// Example:
//
//	vault.WithTLSConfig(&api.TLSConfig{
//	    CACert: "/path/to/ca.crt",
//	    Insecure: false,
//	})
func WithTLSConfig(cfg *vaultapi.TLSConfig) Option {
	return func(c *resolverConfig) {
		c.tlsConfig = cfg
	}
}

// WithKubernetesAuth configures Kubernetes authentication.
// This is the recommended method for applications running in Kubernetes.
//
// Parameters:
//   - role: The Vault role to authenticate as
//   - jwtPath: Path to the service account token (typically /var/run/secrets/kubernetes.io/serviceaccount/token)
//
// Example:
//
//	vault.WithKubernetesAuth("my-app-role", "/var/run/secrets/kubernetes.io/serviceaccount/token")
func WithKubernetesAuth(role, jwtPath string) Option {
	return func(c *resolverConfig) {
		c.authMethod = &kubernetesAuth{
			role:    role,
			jwtPath: jwtPath,
		}
	}
}

// WithAppRole configures AppRole authentication.
// AppRole is designed for machine-to-machine authentication.
//
// Parameters:
//   - roleID: The AppRole role ID
//   - secretID: The AppRole secret ID
//
// Example:
//
//	vault.WithAppRole(os.Getenv("VAULT_ROLE_ID"), os.Getenv("VAULT_SECRET_ID"))
func WithAppRole(roleID, secretID string) Option {
	return func(c *resolverConfig) {
		c.authMethod = &appRoleAuth{
			roleID:   roleID,
			secretID: secretID,
		}
	}
}

// kubernetesAuth implements Kubernetes authentication method.
type kubernetesAuth struct {
	role    string
	jwtPath string
	mount   string // Optional, defaults to "kubernetes"
}

// Login authenticates using Kubernetes service account token.
func (k *kubernetesAuth) Login(ctx context.Context, client *vaultapi.Client) (string, error) {
	// Read the JWT from the service account token file
	jwt, err := os.ReadFile(k.jwtPath)
	if err != nil {
		return "", fmt.Errorf("failed to read kubernetes JWT from %q: %w", k.jwtPath, err)
	}

	mount := k.mount
	if mount == "" {
		mount = "kubernetes"
	}

	// Perform login
	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := client.Logical().WriteWithContext(ctx, path, map[string]any{
		"role": k.role,
		"jwt":  string(jwt),
	})
	if err != nil {
		return "", fmt.Errorf("kubernetes auth failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return "", errors.New("kubernetes auth returned no token")
	}

	return secret.Auth.ClientToken, nil
}

// appRoleAuth implements AppRole authentication method.
type appRoleAuth struct {
	roleID   string
	secretID string
	mount    string // Optional, defaults to "approle"
}

// Login authenticates using AppRole credentials.
func (a *appRoleAuth) Login(ctx context.Context, client *vaultapi.Client) (string, error) {
	mount := a.mount
	if mount == "" {
		mount = "approle"
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := client.Logical().WriteWithContext(ctx, path, map[string]any{
		"role_id":   a.roleID,
		"secret_id": a.secretID,
	})
	if err != nil {
		return "", fmt.Errorf("approle auth failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return "", errors.New("approle auth returned no token")
	}

	return secret.Auth.ClientToken, nil
}

// WithKubernetesAuthMount configures Kubernetes authentication with a custom mount path.
// Use this if your Kubernetes auth method is mounted at a non-default path.
//
// Example:
//
//	vault.WithKubernetesAuthMount("my-k8s", "my-app-role", "/var/run/secrets/kubernetes.io/serviceaccount/token")
func WithKubernetesAuthMount(mount, role, jwtPath string) Option {
	return func(c *resolverConfig) {
		c.authMethod = &kubernetesAuth{
			mount:   mount,
			role:    role,
			jwtPath: jwtPath,
		}
	}
}

// WithAppRoleMount configures AppRole authentication with a custom mount path.
// Use this if your AppRole auth method is mounted at a non-default path.
//
// Example:
//
//	vault.WithAppRoleMount("my-approle", roleID, secretID)
func WithAppRoleMount(mount, roleID, secretID string) Option {
	return func(c *resolverConfig) {
		c.authMethod = &appRoleAuth{
			mount:    mount,
			roleID:   roleID,
			secretID: secretID,
		}
	}
}
