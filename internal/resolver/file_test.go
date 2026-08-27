package resolver

import (
	"errors"
	"path"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &openRecordingFs{}
			resolver := NewFileResolver(fs)

			_, err := resolver.Resolve(t.Context(), tt.uri)
			assert.ErrorIs(t, err, errors.ErrUnsupported)
			// Clean normalizes the "./" prefix the resolver adds for relative paths.
			assert.Equal(t, tt.want, path.Clean(fs.openedPath))
		})
	}
}
