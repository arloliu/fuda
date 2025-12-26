package resolver

import (
	"context"
	"fmt"
	"net/url"
	"os"
)

// FileResolver resolves references using the file:// scheme.
type FileResolver struct{}

// NewFileResolver creates a new FileResolver.
func NewFileResolver() *FileResolver {
	return &FileResolver{}
}

// Resolve reads the file at the given URI.
// Supports both file://path and file:///path formats.
func (r *FileResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI %q: %w", uri, err)
	}

	if u.Scheme != "file" {
		return nil, fmt.Errorf("unsupported scheme for file resolver: %s", u.Scheme)
	}

	// Handle both file://path (host=path, path="") and file:///path (host="", path="/path")
	// The standard file URI format is file:///absolute/path or file://host/path
	// For convenience, we also support file://relative/path where the path is treated as Host
	path := u.Path
	if path == "" && u.Host != "" {
		// file://relative/path format - Host contains the path
		path = u.Host + u.Path
	}

	// Check context before reading
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return os.ReadFile(path)
}
