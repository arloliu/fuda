package fuda

import "context"

// RefResolver is an interface for resolving references.
// It is used to mock reference resolution in tests or provide custom resolution logic.
// Implementations MUST be safe for concurrent use by multiple goroutines.
type RefResolver interface {
	// Resolve returns the content referenced by the uri.
	Resolve(ctx context.Context, uri string) ([]byte, error)
}
