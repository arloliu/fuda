package resolver

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvResolver_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		envVal  string
		uri     string
		want    []byte
		wantErr bool
	}{
		{
			name:    "Basic resolution",
			envVar:  "TEST_VAR_BASIC",
			envVal:  "some-value",
			uri:     "env://TEST_VAR_BASIC",
			want:    []byte("some-value"),
			wantErr: false,
		},
		{
			name:    "Resolution with triple slash",
			envVar:  "TEST_VAR_TRIPLE",
			envVal:  "another-value",
			uri:     "env:///TEST_VAR_TRIPLE",
			want:    []byte("another-value"),
			wantErr: false,
		},
		{
			name: "Unset variable",
			// No env var set
			uri:     "env://TEST_UNSET_VAR",
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "Empty variable",
			envVar:  "TEST_EMPTY_VAR",
			envVal:  "",
			uri:     "env://TEST_EMPTY_VAR",
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "Empty variable name",
			uri:     "env://",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Wrong scheme",
			uri:     "file://test",
			want:    nil,
			wantErr: true,
		},
	}

	r := NewEnvResolver()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
			}

			got, err := r.Resolve(ctx, tt.uri)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
