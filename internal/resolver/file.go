package resolver

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
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

// isASCIILetter reports whether b is an ASCII letter (A-Z or a-z).
func isASCIILetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// isWindowsDriveAuthority reports whether rest begins with a Windows drive
// letter followed by a colon, as in the "C:" of "file://C:/x" or
// "file://C:\x". Without special handling, url.Parse would otherwise consume
// the drive letter as a URI authority and fail trying to parse its ":" as a
// port separator.
func isWindowsDriveAuthority(rest string) bool {
	return len(rest) >= 2 && isASCIILetter(rest[0]) && rest[1] == ':'
}

// toFilesystemPath converts a parsed URL absolute path to a filesystem path.
// Windows drive-form paths such as "/C:/x" have their leading slash stripped
// (producing "C:/x"); the result is then passed through filepath.FromSlash
// to convert to the OS-specific separator.
// On POSIX both steps are no-ops, so "/absolute/path" is returned unchanged.
func toFilesystemPath(p string) string {
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && isASCIILetter(p[1]) {
		p = p[1:]
	}

	return filepath.FromSlash(p)
}

// Resolve reads the file at the given URI.
//
// Three forms are supported:
//   - file:///absolute/path reads an absolute path.
//   - file://relative/path reads a path relative to the process working directory.
//   - file://C:/x, file://C:\x, and file:///C:/x read a Windows drive-form
//     absolute path; the drive letter is treated as part of the path rather
//     than a URI authority.
//
// The file:// prefix is matched case-insensitively, so FILE://relative/path
// and file://relative/path behave the same way.
//
// URI authorities (file://host/path) are not supported; the authority segment
// is interpreted as the first component of a relative path instead.
func (r *FileResolver) Resolve(ctx context.Context, uri string) ([]byte, error) {
	originalURI := uri

	// Rewrite file://relative/path to file://localhost/relative/path so that
	// url.Parse treats the whole path as a path instead of parsing its first
	// component as an RFC 3986 authority.
	// Rewrite a Windows drive-form authority such as file://C:/x to
	// file:///C:/x for the same reason.
	const nonstandardPrefix = "file://"
	isNonstandardRelative := false
	if len(uri) > len(nonstandardPrefix) && strings.EqualFold(uri[:len(nonstandardPrefix)], nonstandardPrefix) {
		rest := uri[len(nonstandardPrefix):]
		switch {
		case rest[0] == '/':
			// Already a standard authority-less absolute URI; leave as-is.
		case isWindowsDriveAuthority(rest):
			uri = nonstandardPrefix + "/" + rest
		default:
			isNonstandardRelative = true
			uri = nonstandardPrefix + "localhost/" + rest
		}
	}

	u, err := url.Parse(uri)
	if err != nil {
		// url.Error.Error() re-quotes uri, which may be the internally
		// rewritten form above; unwrap to the underlying cause so the
		// reported message only ever shows the caller's original URI.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}

		return nil, fmt.Errorf("invalid URI %q: %w", originalURI, err)
	}

	if u.Scheme != "file" {
		return nil, fmt.Errorf("unsupported scheme for file resolver: %s", u.Scheme)
	}

	var path string
	switch {
	case isNonstandardRelative:
		// file://relative/path => file://localhost/relative/path => /relative/path => ./relative/path
		path = "." + u.Path
	default:
		path = toFilesystemPath(u.Path)
	}

	// Check context before reading
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return afero.ReadFile(r.fs, path)
}
