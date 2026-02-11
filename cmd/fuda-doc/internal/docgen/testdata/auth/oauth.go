// Package auth provides configuration types for authentication providers.
package auth

import "time"

// OAuthConfig configures OAuth 2.0 / OpenID Connect authentication.
type OAuthConfig struct {
	// Issuer is the OIDC issuer URL used for auto-discovery.
	//
	// The well-known configuration will be fetched from:
	//   {Issuer}/.well-known/openid-configuration
	//
	// Example:
	//   https://accounts.google.com
	//   https://login.microsoftonline.com/{tenant}/v2.0
	Issuer string `yaml:"issuer" env:"OAUTH_ISSUER" validate:"required,url"`

	// ClientID is the OAuth client identifier.
	ClientID string `yaml:"client_id" env:"OAUTH_CLIENT_ID" validate:"required"`

	// ClientSecret is the OAuth client secret.
	// Should be provided via environment variable or secret store.
	ClientSecret string `yaml:"client_secret" env:"OAUTH_CLIENT_SECRET" validate:"required"`

	// Scopes lists the OAuth scopes to request during authorization.
	Scopes []string `yaml:"scopes" default:"openid,profile,email"`

	// TokenExpiry controls how long access tokens are cached locally.
	TokenExpiry time.Duration `yaml:"token_expiry" default:"1h"`

	// Providers holds named provider-specific overrides.
	// Keys are provider names (e.g., "google", "github").
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig stores per-provider OAuth overrides.
type ProviderConfig struct {
	// AuthURL overrides the authorization endpoint.
	AuthURL string `yaml:"auth_url,omitempty"`

	// TokenURL overrides the token endpoint.
	TokenURL string `yaml:"token_url,omitempty"`

	// UserInfoURL overrides the userinfo endpoint.
	UserInfoURL string `yaml:"userinfo_url,omitempty"`

	// ExtraScopes are additional scopes to request for this provider.
	ExtraScopes []string `yaml:"extra_scopes,omitempty"`
}
