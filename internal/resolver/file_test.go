package resolver

import (
	"errors"
	"os"
	"path"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type openRecordingFs struct {
	openedPath string
}

var _ afero.Fs = &openRecordingFs{}

func (openRecordingFs) Create(string) (afero.File, error) {
	return nil, errors.ErrUnsupported
}

func (openRecordingFs) Mkdir(string, os.FileMode) error {
	return errors.ErrUnsupported
}

func (openRecordingFs) MkdirAll(string, os.FileMode) error {
	return errors.ErrUnsupported
}

func (f *openRecordingFs) Open(name string) (afero.File, error) {
	f.openedPath = name
	return nil, errors.ErrUnsupported
}

func (openRecordingFs) OpenFile(string, int, os.FileMode) (afero.File, error) {
	return nil, errors.ErrUnsupported
}

func (openRecordingFs) Remove(string) error {
	return errors.ErrUnsupported
}

func (openRecordingFs) RemoveAll(string) error {
	return errors.ErrUnsupported
}

func (openRecordingFs) Rename(string, string) error {
	return errors.ErrUnsupported
}

func (openRecordingFs) Stat(string) (os.FileInfo, error) {
	return nil, errors.ErrUnsupported
}

func (openRecordingFs) Name() string {
	return "openRecordingFs"
}

func (openRecordingFs) Chmod(string, os.FileMode) error {
	return errors.ErrUnsupported
}

func (openRecordingFs) Chown(string, int, int) error {
	return errors.ErrUnsupported
}

func (openRecordingFs) Chtimes(string, time.Time, time.Time) error {
	return errors.ErrUnsupported
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &openRecordingFs{}
			resolver := NewFileResolver(fs)

			_, err := resolver.Resolve(t.Context(), tt.uri)
			assert.ErrorIs(t, errors.ErrUnsupported, err)
			assert.Equal(t, tt.want, path.Clean(fs.openedPath))
		})
	}
}
