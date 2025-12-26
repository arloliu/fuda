package types_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/arloliu/fuda/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Nested struct {
	Val string
}

func TestConvert(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		target    any
		expected  any
		shouldErr bool
	}{
		// String
		{"string", "hello", new(string), "hello", false},

		// Integers
		{"int", "123", new(int), 123, false},
		{"int8", "123", new(int8), int8(123), false},
		{"int16", "123", new(int16), int16(123), false},
		{"int32", "123", new(int32), int32(123), false},
		{"int64", "123", new(int64), int64(123), false},
		{"uint", "123", new(uint), uint(123), false},

		// Floats
		{"float64", "123.456", new(float64), 123.456, false},

		// Booleans
		{"bool true", "true", new(bool), true, false},
		{"bool false", "false", new(bool), false, false},

		// Duration
		{"duration", "10s", new(time.Duration), 10 * time.Second, false},

		// Slices
		{"slice string", "a,b,c", new([]string), []string{"a", "b", "c"}, false},
		{"slice int", "1,2,3", new([]int), []int{1, 2, 3}, false},

		// Maps
		{"map string", "key:val,key2:val2", new(map[string]string), map[string]string{"key": "val", "key2": "val2"}, false},
		{"map int", "key:1,key2:2", new(map[string]int), map[string]int{"key": 1, "key2": 2}, false},

		// Struct (JSON)
		{"struct", `{"Val":"test"}`, new(Nested), Nested{Val: "test"}, false},

		// Pointer
		{"pointer", "123", new(*int), func() *int {
			i := 123

			return &i
		}(), false},

		// Errors
		{"invalid int", "abc", new(int), nil, true},
		{"invalid bool", "notbool", new(bool), nil, true},
		{"invalid map", "invalid", new(map[string]string), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.target).Elem()
			err := types.Convert(tt.input, val)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if val.Kind() == reflect.Ptr && tt.name == "pointer" {
					// Compare dereferenced value for pointers
					require.Equal(t, *(tt.expected.(*int)), val.Elem().Interface()) //nolint:errcheck
				} else {
					assert.Equal(t, tt.expected, val.Interface())
				}
			}
		})
	}
}
