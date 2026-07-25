package resolver

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/afero"
)

// FileResolver resolves references using the file:// scheme.
type FileResolver struct {
	fs afero.Fs
}

// NewFileResolver creates a new FileResolver.
// If fs is nil, the OS filesystem is used.
func NewFileResolver(fs afero.Fs) *FileResolver {
	if fs == nil {
		fs = afero.NewOsFs()
	}

	return &FileResolver{fs: fs}
}

// Resolve reads the file at the given URI.
// Supports both file://path and file:///path formats.
func (r *FileResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	// Handle both file://path (authority=path, path="") and file:///path (authority="", path="/path")
	// The standard file URI format is file:///absolute/path or file://authority/path
	// We do not support authority, for convenience, we interpret as file://relative/path instead.

	// Map file://relative/path to file://localhost/relative/path to fix semantic to parse as RFC 3986 URI
	const nonstandardRelativePrefix = "file://"
	isNonstandardRelative := false
	if strings.HasPrefix(uri, nonstandardRelativePrefix) &&
		len(uri) > len(nonstandardRelativePrefix) && uri[len(nonstandardRelativePrefix)] != '/' {
		isNonstandardRelative = true
		uri = "file://localhost/" + uri[len(nonstandardRelativePrefix):]
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI %q: %w", uri, err)
	}

	if u.Scheme != "file" {
		return nil, fmt.Errorf("unsupported scheme for file resolver: %s", u.Scheme)
	}

	path := u.Path
	// file://relative/path => file://localhost/relative/path => /relative/path => ./relative/path
	if isNonstandardRelative {
		path = "." + u.Path
	}

	// Check context before reading
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return afero.ReadFile(r.fs, path)
}
