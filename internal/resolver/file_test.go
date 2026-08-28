package resolver

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// openRecordingFs records the path passed to Open and fails every call.
// The embedded nil afero.Fs panics on any other method, which flags
// unexpected filesystem access loudly.
type openRecordingFs struct {
	afero.Fs

	openedPath string
}

var _ afero.Fs = (*openRecordingFs)(nil)

func (f *openRecordingFs) Open(name string) (afero.File, error) {
	f.openedPath = name
	return nil, errors.ErrUnsupported
}

func TestFileResolver_Resolve(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "Absolute path",
			uri:  "file:///absolute/path",
			want: "/absolute/path",
		},
		{
			name: "Relative path",
			uri:  "file://relative/path",
			want: "relative/path",
		},
		{
			name: "Relative path with ./",
			uri:  "file://./relative/path",
			want: "relative/path",
		},
		{
			name: "Relative path with authority-like first component",
			uri:  "file://userinfo@host:123/relative/path",
			want: "userinfo@host:123/relative/path",
		},
		{
			name: "Relative path with first component invalid for authority",
			uri:  "file://userinfo@host:notport/relative/path",
			want: "userinfo@host:notport/relative/path",
		},
		{
			name: "Relative path with parent traversal",
			uri:  "file://../parent/path",
			want: "../parent/path",
		},
		{
			name: "Relative path with percent-encoded space",
			uri:  "file://my%20dir/my%20file",
			want: "my dir/my file",
		},
		{
			name: "Absolute path with percent-encoded space",
			uri:  "file:///my%20dir/my%20file",
			want: "/my dir/my file",
		},
		{
			name: "Uppercase scheme with relative path",
			uri:  "FILE://relative/path",
			want: "relative/path",
		},
		{
			name: "Scheme only with no authority or path",
			uri:  "file://",
			want: ".",
		},
		{
			name: "Current directory only",
			uri:  "file://.",
			want: ".",
		},
		{
			name: "Parent directory only",
			uri:  "file://..",
			want: "..",
		},
		{
			name: "Root absolute path",
			uri:  "file:///",
			want: "/",
		},
		{
			name: "Double leading slash absolute path collapses to one",
			uri:  "file:////x",
			want: "/x",
		},
		{
			name: "Relative path with trailing slash",
			uri:  "file://relative/path/",
			want: "relative/path",
		},
		{
			name: "Relative path with percent-encoded slash decodes as a separator",
			uri:  "file://my%2Ffile",
			want: "my/file",
		},
		{
			name: "Relative path with raw fragment delimiter drops the suffix",
			uri:  "file://a#b",
			want: "a",
		},
		{
			name: "Absolute Windows drive path",
			uri:  "file:///C:/x",
			want: "C:/x",
		},
		{
			name: "Authority-form Windows drive path",
			uri:  "file://C:/x",
			want: "C:/x",
		},
		{
			name: "Authority-form Windows drive path with backslash",
			uri:  `file://C:\x`,
			want: `C:\x`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &openRecordingFs{}
			resolver := NewFileResolver(fs)

			_, err := resolver.Resolve(t.Context(), tt.uri)
			assert.ErrorIs(t, err, errors.ErrUnsupported)
			// filepath.Clean normalizes the "./" prefix the resolver adds for
			// relative paths and the redundant separators in some edge cases.
			// tt.want is authored with forward slashes; filepath.FromSlash
			// converts it to the OS-specific form so the comparison holds on
			// both POSIX and Windows.
			assert.Equal(t, filepath.FromSlash(tt.want), filepath.Clean(fs.openedPath))
		})
	}
}

func TestFileResolver_Resolve_InvalidEscapeReportsOriginalURI(t *testing.T) {
	fs := &openRecordingFs{}
	resolver := NewFileResolver(fs)

	_, err := resolver.Resolve(t.Context(), "file://bad%zz")
	assert.ErrorContains(t, err, `"file://bad%zz"`)
	assert.NotContains(t, err.Error(), "localhost")
}
