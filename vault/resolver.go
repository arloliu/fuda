// Package vault provides a HashiCorp Vault resolver for fuda.
//
// This package implements [fuda.RefResolver] to fetch secrets from Vault
// using the vault:// URI scheme. It supports multiple authentication methods
// including Token, Kubernetes, and AppRole.
//
// Basic usage:
//
//	resolver, err := vault.NewResolver(
//	    vault.WithAddress("https://vault.example.com:8200"),
//	    vault.WithToken(os.Getenv("VAULT_TOKEN")),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithRefResolver(resolver).
//	    Build()
//
// # URI Format
//
// The vault resolver uses the following URI format:
//
//	vault:///<mount>/<path>#<field>
//
// Examples:
//   - vault:///secret/data/myapp#password (KV v2)
//   - vault:///kv/myapp#api_key (KV v1)
//   - vault:///database/creds/readonly#username (Dynamic secrets)
//
// # Authentication Methods
//
// Token authentication:
//
//	vault.WithToken(os.Getenv("VAULT_TOKEN"))
//
// Kubernetes authentication (for pods running in K8s):
//
//	vault.WithKubernetesAuth("my-role", "/var/run/secrets/kubernetes.io/serviceaccount/token")
//
// AppRole authentication:
//
//	vault.WithAppRole(roleID, secretID)
package vault

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// Resolver implements fuda.RefResolver for HashiCorp Vault.
// It resolves vault:// URIs by fetching secrets from a Vault server.
type Resolver struct {
	client    *vaultapi.Client
	config    *resolverConfig
	authDone  bool
	namespace string
}

// resolverConfig holds internal configuration for the resolver.
type resolverConfig struct {
	address    string
	token      string
	namespace  string
	authMethod authMethod
	tlsConfig  *vaultapi.TLSConfig
}

// authMethod represents a Vault authentication method.
type authMethod interface {
	// Login performs authentication and returns a token.
	Login(ctx context.Context, client *vaultapi.Client) (string, error)
}

// NewResolver creates a new Vault resolver with the given options.
//
// At minimum, you must provide an address and an authentication method:
//
//	resolver, err := vault.NewResolver(
//	    vault.WithAddress("https://vault.example.com:8200"),
//	    vault.WithToken(os.Getenv("VAULT_TOKEN")),
//	)
//
// Available options:
//   - [WithAddress] - Vault server address (required)
//   - [WithToken] - Token authentication
//   - [WithKubernetesAuth] - Kubernetes authentication
//   - [WithAppRole] - AppRole authentication
//   - [WithNamespace] - Vault namespace (Enterprise)
//   - [WithTLSConfig] - Custom TLS configuration
func NewResolver(opts ...Option) (*Resolver, error) {
	cfg := &resolverConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.address == "" {
		return nil, errors.New("vault address is required: use WithAddress()")
	}

	// Create Vault client config
	vaultCfg := vaultapi.DefaultConfig()
	vaultCfg.Address = cfg.address

	if cfg.tlsConfig != nil {
		if err := vaultCfg.ConfigureTLS(cfg.tlsConfig); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
	}

	client, err := vaultapi.NewClient(vaultCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Set namespace if provided (Enterprise feature)
	if cfg.namespace != "" {
		client.SetNamespace(cfg.namespace)
	}

	// Set token directly if provided
	if cfg.token != "" {
		client.SetToken(cfg.token)
	}

	return &Resolver{
		client:    client,
		config:    cfg,
		namespace: cfg.namespace,
	}, nil
}

// Resolve fetches the secret value from Vault for the given URI.
//
// URI format: vault:///<mount>/<path>#<field>
//
// The resolver automatically handles both KV v1 and KV v2 secrets engines.
// For KV v2, the data is extracted from the nested "data" field.
func (r *Resolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	// Authenticate if needed (lazy authentication)
	if err := r.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("vault authentication failed: %w", err)
	}

	// Parse URI
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid vault URI %q: %w", uri, err)
	}

	if u.Scheme != "vault" {
		return nil, fmt.Errorf("unsupported scheme %q: expected vault://", u.Scheme)
	}

	// Extract path and field from URI
	// vault:///secret/data/myapp#password
	// Path: /secret/data/myapp, Fragment: password
	path := strings.TrimPrefix(u.Path, "/")
	field := u.Fragment

	if path == "" {
		return nil, fmt.Errorf("vault URI missing path: %s", uri)
	}
	if field == "" {
		return nil, fmt.Errorf("vault URI missing field (fragment): %s", uri)
	}

	// Check context before making request
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Read secret from Vault
	secret, err := r.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault secret at %q: %w", path, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("vault secret not found at %q", path)
	}

	// Extract the field value
	value, err := r.extractField(secret.Data, field, path)
	if err != nil {
		return nil, err
	}

	return []byte(value), nil
}

// ensureAuthenticated performs lazy authentication if an auth method is configured.
func (r *Resolver) ensureAuthenticated(ctx context.Context) error {
	// Skip if already authenticated or using direct token
	if r.authDone || r.config.token != "" {
		return nil
	}

	if r.config.authMethod == nil {
		return errors.New("no authentication method configured: use WithToken(), WithKubernetesAuth(), or WithAppRole()")
	}

	token, err := r.config.authMethod.Login(ctx, r.client)
	if err != nil {
		return err
	}

	r.client.SetToken(token)
	r.authDone = true

	return nil
}

// extractField extracts a field value from Vault secret data.
// It handles both KV v1 (flat) and KV v2 (nested under "data") formats.
func (r *Resolver) extractField(data map[string]any, field, path string) (string, error) {
	// Check for KV v2 format (data nested under "data" key)
	if nestedData, ok := data["data"].(map[string]any); ok {
		if value, ok := nestedData[field]; ok {
			return r.valueToString(value, field, path)
		}
		// If field not found in nested data, check top-level too
	}

	// Check top-level (KV v1 format or other secret engines)
	if value, ok := data[field]; ok {
		return r.valueToString(value, field, path)
	}

	return "", fmt.Errorf("field %q not found in vault secret at %q", field, path)
}

// valueToString converts a secret field value to a string.
func (r *Resolver) valueToString(value any, field, path string) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("vault field %q at %q is not a string (got %T)", field, path, value)
	}
}

// Client returns the underlying Vault API client for advanced usage.
// This allows users to perform operations not covered by the resolver interface.
func (r *Resolver) Client() *vaultapi.Client {
	return r.client
}
