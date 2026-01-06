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

// mockResolver for testing ref tag resolution
type mockByteResolver struct {
	data map[string][]byte
}

func (m *mockByteResolver) Resolve(_ context.Context, uri string) ([]byte, error) {
	if val, ok := m.data[uri]; ok {
		return val, nil
	}
	return nil, os.ErrNotExist
}

// Test structs for []byte ref tag tests
type ByteSliceRefStruct struct {
	Data []byte `ref:"file://test-data"`
}

type ByteSliceRefFromStruct struct {
	DataRef string `default:"file://test-data"`
	Data    []byte `refFrom:"DataRef"`
}

type ByteSliceRefFromExplicitStruct struct {
	DataRef string
	Data    []byte `refFrom:"DataRef"`
}

type ByteSliceDefaultStruct struct {
	Data []byte `default:"default-content"`
}

func TestProcessRef_ByteSlice(t *testing.T) {
	ctx := context.Background()
	resolver := &mockByteResolver{
		data: map[string][]byte{
			"file://test-data": []byte("binary-content-here"),
		},
	}

	t.Run("ref tag with []byte", func(t *testing.T) {
		s := ByteSliceRefStruct{}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("Data")
		val := v.FieldByName("Data")

		resolved, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.True(t, resolved)
		assert.Equal(t, []byte("binary-content-here"), s.Data)
	})

	t.Run("refFrom tag with []byte", func(t *testing.T) {
		s := ByteSliceRefFromExplicitStruct{
			DataRef: "file://test-data",
		}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("Data")
		val := v.FieldByName("Data")

		resolved, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.True(t, resolved)
		assert.Equal(t, []byte("binary-content-here"), s.Data)
	})

	t.Run("refFrom tag with default peeking", func(t *testing.T) {
		// DataRef is empty, but has default:"file://test-data"
		s := ByteSliceRefFromStruct{}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("Data")
		val := v.FieldByName("Data")

		resolved, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.True(t, resolved)
		assert.Equal(t, []byte("binary-content-here"), s.Data)
	})

	t.Run("binary content preservation", func(t *testing.T) {
		// Test with actual binary data including non-UTF8 bytes
		binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0x80, 0x7F}
		resolver := &mockByteResolver{
			data: map[string][]byte{
				"file://binary-file": binaryData,
			},
		}

		type Config struct {
			BinaryData []byte `ref:"file://binary-file"`
		}

		s := Config{}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("BinaryData")
		val := v.FieldByName("BinaryData")

		resolved, err := tags.ProcessRef(ctx, field, val, v, resolver, "", nil)
		require.NoError(t, err)
		assert.True(t, resolved)
		assert.Equal(t, binaryData, s.BinaryData)
	})
}

func TestProcessDefault_ByteSlice(t *testing.T) {
	t.Run("default tag with []byte", func(t *testing.T) {
		s := ByteSliceDefaultStruct{}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("Data")
		val := v.FieldByName("Data")

		err := tags.ProcessDefault(field, val)
		require.NoError(t, err)
		assert.Equal(t, []byte("default-content"), s.Data)
	})

	t.Run("default tag with []byte skips non-zero", func(t *testing.T) {
		s := ByteSliceDefaultStruct{
			Data: []byte("existing"),
		}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("Data")
		val := v.FieldByName("Data")

		err := tags.ProcessDefault(field, val)
		require.NoError(t, err)
		assert.Equal(t, []byte("existing"), s.Data, "Should not overwrite existing value")
	})
}

func TestProcessEnv_ByteSlice(t *testing.T) {
	type EnvByteStruct struct {
		Data []byte `env:"TEST_BYTE_DATA"`
	}

	t.Run("env tag with []byte", func(t *testing.T) {
		t.Setenv("TEST_BYTE_DATA", "env-binary-content")

		s := EnvByteStruct{}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("Data")
		val := v.FieldByName("Data")

		err := tags.ProcessEnv(field, val, "")
		require.NoError(t, err)
		assert.Equal(t, []byte("env-binary-content"), s.Data)
	})

	t.Run("env tag with []byte and prefix", func(t *testing.T) {
		t.Setenv("APP_TEST_BYTE_DATA", "prefixed-env-content")

		s := EnvByteStruct{}
		v := reflect.ValueOf(&s).Elem()
		typ := v.Type()

		field, _ := typ.FieldByName("Data")
		val := v.FieldByName("Data")

		err := tags.ProcessEnv(field, val, "APP_")
		require.NoError(t, err)
		assert.Equal(t, []byte("prefixed-env-content"), s.Data)
	})
}
