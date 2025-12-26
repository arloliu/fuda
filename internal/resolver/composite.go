package resolver

import (
	"context"
	"fmt"
	"strings"
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
func New() *CompositeResolver {
	cr := &CompositeResolver{
		resolvers: make(map[string]SubResolver),
	}
	cr.Register("file", NewFileResolver())

	httpResolver := NewHTTPResolver()
	cr.Register("http", httpResolver)
	cr.Register("https", httpResolver)

	return cr
}

// Register registers a sub-resolver for a given scheme.
func (r *CompositeResolver) Register(scheme string, resolver SubResolver) {
	r.resolvers[scheme] = resolver
}

// Resolve delegates resolution to the appropriate sub-resolver.
func (r *CompositeResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid uri format: %s", uri)
	}
	scheme := parts[0]

	resolver, ok := r.resolvers[scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %s", scheme)
	}

	return resolver.Resolve(ctx, uri)
}
