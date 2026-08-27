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
//
// Two forms are supported:
//   - file:///absolute/path reads an absolute path.
//   - file://relative/path reads a path relative to the process working directory.
//
// URI authorities (file://host/path) are not supported; the authority segment
// is interpreted as the first component of a relative path instead.
func (r *FileResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	// Rewrite file://relative/path to file://localhost/relative/path so that
	// url.Parse treats the whole path as a path instead of parsing its first
	// component as an RFC 3986 authority.
	const nonstandardRelativePrefix = "file://"
	isNonstandardRelative := false
	if len(uri) > len(nonstandardRelativePrefix) &&
		strings.EqualFold(uri[:len(nonstandardRelativePrefix)], nonstandardRelativePrefix) &&
		uri[len(nonstandardRelativePrefix)] != '/' {
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
