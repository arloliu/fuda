package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/afero"
)

type SubResolver interface {
	// Resolve returns the content referenced by the uri.
	Resolve(ctx context.Context, uri string) ([]byte, error)
}

// CompositeResolver delegates resolution to sub-resolvers based on scheme.
type CompositeResolver struct {
	resolvers map[string]SubResolver
}

// New creates a new CompositeResolver with default sub-resolvers.
// If fs is nil, the OS filesystem is used for file:// resolution.
func New(fs afero.Fs) *CompositeResolver {
	cr := &CompositeResolver{
		resolvers: make(map[string]SubResolver),
	}
	cr.Register("file", NewFileResolver(fs))

	httpResolver := NewHTTPResolver()
	cr.Register("http", httpResolver)
	cr.Register("https", httpResolver)
	cr.Register("env", NewEnvResolver())

	return cr
}

// Register registers a sub-resolver for a given scheme.
// The scheme is stored in lowercase so that Resolve, which also lowercases
// the scheme it extracts from a URI, matches it regardless of the case used
// at registration or call time.
func (r *CompositeResolver) Register(scheme string, resolver SubResolver) {
	r.resolvers[strings.ToLower(scheme)] = resolver
}

// Resolve delegates resolution to the appropriate sub-resolver.
func (r *CompositeResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid uri format: %s", uri)
	}
	scheme := parts[0]

	resolver, ok := r.resolvers[strings.ToLower(scheme)]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %s", scheme)
	}

	return resolver.Resolve(ctx, uri)
}
