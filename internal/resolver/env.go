package resolver

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// EnvResolver resolves references using the env:// scheme.
type EnvResolver struct{}

// NewEnvResolver creates a new EnvResolver.
func NewEnvResolver() *EnvResolver {
	return &EnvResolver{}
}

// Resolve reads the environment variable specified in the URI.
// URI format: env://VAR_NAME
// Returns empty []byte if the variable is not set.
func (r *EnvResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	// 1. Parse URI
	// We handle simple env://VAR cases manually to avoid url.Parse issues with some chars,
	// but using url.Parse is safer for scheme validation.
	// However, env vars can contain characters that might confuse url.Parse if not encoded.
	// For simplicity and consistency with file resolver, let's try standard parsing first.
	// But commonly env://VAR_NAME is used. host=VAR_NAME.

	if !strings.HasPrefix(uri, "env://") {
		return nil, fmt.Errorf("unsupported scheme for env resolver: %s", uri)
	}

	// Extract variable name.
	// Valid formats:
	// env://VAR_NAME      -> host=VAR_NAME
	// env:///VAR_NAME     -> path=/VAR_NAME (strip leading /)

	// We do "loose" parsing to support potentially weird env var names
	trimmed := strings.TrimPrefix(uri, "env://")

	// If it starts with /, treat as env:///VAR_NAME -> VAR_NAME
	varName := strings.TrimPrefix(trimmed, "/")

	// If we want to be strict about URL parsing:
	u, err := url.Parse(uri)
	if err == nil && u.Scheme == "env" {
		if u.Host != "" {
			varName = u.Host
		} else if u.Path != "" {
			varName = strings.TrimPrefix(u.Path, "/")
		}
	}

	if varName == "" {
		return nil, fmt.Errorf("empty environment variable name in URI: %s", uri)
	}

	val, ok := os.LookupEnv(varName)
	if !ok {
		// Variable not set - return ErrNotExist to signal "not found" for fallback chain
		return nil, os.ErrNotExist
	}

	// Variable is set (even if empty) - return its value
	return []byte(val), nil
}
