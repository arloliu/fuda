package tags_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/arloliu/fuda/internal/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Field       string `default:"default_val"`
	Empty       string
	EnvField    string `env:"TEST_TAG_ENV"`
	RefField    string `ref:"file://test_ref"`
	RefFrom     string `refFrom:"RefPath"`
	RefPath     string `default:"test_ref"`
	RefPathBare string
}

func TestProcessDefault(t *testing.T) {
	s := TestStruct{}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()

	t.Run("apply default", func(t *testing.T) {
		field, _ := typ.FieldByName("Field")
		val := v.FieldByName("Field")
		err := tags.ProcessDefault(field, val)
		require.NoError(t, err)
		assert.Equal(t, "default_val", s.Field)
	})

	t.Run("skip non-zero", func(t *testing.T) {
		s.Field = "existing"
		field, _ := typ.FieldByName("Field")
		val := v.FieldByName("Field")
		err := tags.ProcessDefault(field, val)
		require.NoError(t, err)
		assert.Equal(t, "existing", s.Field)
	})

	t.Run("no default tag", func(t *testing.T) {
		field, _ := typ.FieldByName("Empty")
		val := v.FieldByName("Empty")
		err := tags.ProcessDefault(field, val)
		require.NoError(t, err)
		assert.Equal(t, "", s.Empty)
	})
}

func TestProcessEnv(t *testing.T) {
	s := TestStruct{}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()

	t.Run("apply env", func(t *testing.T) {
		os.Setenv("TEST_TAG_ENV", "env_val")
		defer os.Unsetenv("TEST_TAG_ENV")

		field, _ := typ.FieldByName("EnvField")
		val := v.FieldByName("EnvField")
		err := tags.ProcessEnv(field, val, "")
		require.NoError(t, err)
		assert.Equal(t, "env_val", s.EnvField)
	})

	t.Run("apply env with prefix", func(t *testing.T) {
		os.Setenv("APP_TEST_TAG_ENV", "prefixed_val")
		defer os.Unsetenv("APP_TEST_TAG_ENV")

		field, _ := typ.FieldByName("EnvField")
		val := v.FieldByName("EnvField")
		err := tags.ProcessEnv(field, val, "APP_")
		require.NoError(t, err)
		assert.Equal(t, "prefixed_val", s.EnvField)
	})
}

type mockResolver struct {
	data map[string][]byte
}

func (m *mockResolver) Resolve(_ context.Context, uri string) ([]byte, error) {
	if val, ok := m.data[uri]; ok {
		return val, nil
	}

	return nil, os.ErrNotExist
}

func TestProcessRef(t *testing.T) {
	s := TestStruct{}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()
	resolver := &mockResolver{
		data: map[string][]byte{
			"file://test_ref": []byte("resolved_content"),
		},
	}

	t.Run("ref tag", func(t *testing.T) {
		field, _ := typ.FieldByName("RefField")
		val := v.FieldByName("RefField")
		err := tags.ProcessRef(ctx, field, val, v, resolver)
		require.NoError(t, err)
		assert.Equal(t, "resolved_content", s.RefField)
	})

	t.Run("refFrom tag with default peeking", func(t *testing.T) {
		s.RefFrom = ""
		// RefPath is zero in s, but has `default:"test_ref"` tag
		// RefFrom has `refFrom:"RefPath"`
		// Should look up RefPath (zero) -> check default ("test_ref") -> normalize ("file://test_ref") -> resolve

		field, _ := typ.FieldByName("RefFrom")
		val := v.FieldByName("RefFrom") // RefFrom field

		err := tags.ProcessRef(ctx, field, val, v, resolver)
		require.NoError(t, err)
		assert.Equal(t, "resolved_content", s.RefFrom)
	})

	t.Run("refFrom tag with set value", func(t *testing.T) {
		s.RefFrom = ""
		s.RefPath = "file://test_ref"

		field, _ := typ.FieldByName("RefFrom")
		val := v.FieldByName("RefFrom")

		err := tags.ProcessRef(ctx, field, val, v, resolver)
		require.NoError(t, err)
		assert.Equal(t, "resolved_content", s.RefFrom)
	})
}
