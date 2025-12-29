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
		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
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

		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.Equal(t, "resolved_content", s.RefFrom)
	})

	t.Run("refFrom tag with set value", func(t *testing.T) {
		s.RefFrom = ""
		s.RefPath = "file://test_ref"

		field, _ := typ.FieldByName("RefFrom")
		val := v.FieldByName("RefFrom")

		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.Equal(t, "resolved_content", s.RefFrom)
	})
}

// Test struct for ref template tests
type RefTemplateStruct struct {
	SecretDir string `default:"/etc/secrets"`
	Account   string `default:"admin"`
	Password  string `ref:"file://${.SecretDir}/${.Account}-password"`
}

type RefTemplateNestedDatabase struct {
	CredDir string
}

type RefTemplateNestedStruct struct {
	Database RefTemplateNestedDatabase
	Account  string
	Password string `ref:"file://${.Database.CredDir}/${.Account}-password"`
}

type RefTemplateEnvStruct struct {
	FileName string
	Content  string `ref:"file://${env:TEST_REF_DIR}/${.FileName}"`
}

func TestProcessRef_Template(t *testing.T) {
	ctx := context.Background()

	t.Run("template with field references", func(t *testing.T) {
		s := RefTemplateStruct{
			SecretDir: "/data/secrets",
			Account:   "myuser",
		}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		resolver := &mockResolver{
			data: map[string][]byte{
				"file:///data/secrets/myuser-password": []byte("secret123"),
			},
		}

		field, _ := typ.FieldByName("Password")
		val := v.FieldByName("Password")

		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.Equal(t, "secret123", s.Password)
	})

	t.Run("template with nested field references", func(t *testing.T) {
		s := RefTemplateNestedStruct{
			Account: "testuser",
		}
		s.Database.CredDir = "/var/creds"

		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		resolver := &mockResolver{
			data: map[string][]byte{
				"file:///var/creds/testuser-password": []byte("nestedpass"),
			},
		}

		field, _ := typ.FieldByName("Password")
		val := v.FieldByName("Password")

		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.Equal(t, "nestedpass", s.Password)
	})

	t.Run("template with env function", func(t *testing.T) {
		t.Setenv("TEST_REF_DIR", "/secrets")

		s := RefTemplateEnvStruct{
			FileName: "password.txt",
		}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		resolver := &mockResolver{
			data: map[string][]byte{
				"file:///secrets/password.txt": []byte("envpass"),
			},
		}

		field, _ := typ.FieldByName("Content")
		val := v.FieldByName("Content")

		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.Equal(t, "envpass", s.Content)
	})

	t.Run("template with env function and prefix", func(t *testing.T) {
		t.Setenv("APP_TEST_REF_DIR", "/app/secrets")

		s := RefTemplateEnvStruct{
			FileName: "creds.txt",
		}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		resolver := &mockResolver{
			data: map[string][]byte{
				"file:///app/secrets/creds.txt": []byte("prefixedpass"),
			},
		}

		field, _ := typ.FieldByName("Content")
		val := v.FieldByName("Content")

		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "APP_", nil)
		require.NoError(t, err)
		assert.Equal(t, "prefixedpass", s.Content)
	})

	t.Run("template with missing field uses empty string", func(t *testing.T) {
		// When a referenced field is empty, template should produce empty string
		s := RefTemplateStruct{
			SecretDir: "/etc/secrets",
			Account:   "", // Empty account
		}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		resolver := &mockResolver{
			data: map[string][]byte{
				"file:///etc/secrets/-password": []byte("emptyaccount"),
			},
		}

		field, _ := typ.FieldByName("Password")
		val := v.FieldByName("Password")

		_, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.Equal(t, "emptyaccount", s.Password)
	})
}

func TestRefFromPointerSupport(t *testing.T) {
	type Config struct {
		SourceNil   *string
		SourceEmpty *string
		SourceVal   *string

		SecretNil   string `refFrom:"SourceNil" ref:"file://fallback-nil"`
		SecretEmpty string `refFrom:"SourceEmpty" ref:"file://fallback-empty"`
		SecretVal   string `refFrom:"SourceVal" ref:"file://fallback-val"`
	}

	val := "source-value"
	empty := ""
	s := Config{
		SourceNil:   nil,
		SourceEmpty: &empty,
		SourceVal:   &val,
	}

	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	resolver := &mockResolver{
		data: map[string][]byte{
			"file://fallback-nil":   []byte("fallback-used"),
			"file://fallback-empty": []byte("should-not-be-used"),
			"file://fallback-val":   []byte("should-not-be-used"),
			"file://source-value":   []byte("resolved-from-source"),
		},
	}

	t.Run("nil pointer falls back", func(t *testing.T) {
		field, _ := typ.FieldByName("SecretNil")
		val := v.FieldByName("SecretNil")
		resolved, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.True(t, resolved, "Should resolve from ref tag")
		assert.Equal(t, "fallback-used", s.SecretNil)
	})

	t.Run("empty pointer stops fallback", func(t *testing.T) {
		field, _ := typ.FieldByName("SecretEmpty")
		val := v.FieldByName("SecretEmpty")
		resolved, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.True(t, resolved, "Explicit empty pointer should mark as resolved")
		assert.Equal(t, "", s.SecretEmpty, "Should use empty value from source")
	})

	t.Run("value pointer uses value", func(t *testing.T) {
		field, _ := typ.FieldByName("SecretVal")
		val := v.FieldByName("SecretVal")
		resolved, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.True(t, resolved, "Value pointer should resolve")
		assert.Equal(t, "resolved-from-source", s.SecretVal)
	})
}
