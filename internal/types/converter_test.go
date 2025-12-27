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

		// Duration - standard units (passthrough to time.ParseDuration)
		{"duration seconds", "10s", new(time.Duration), 10 * time.Second, false},
		{"duration minutes", "30m", new(time.Duration), 30 * time.Minute, false},
		{"duration hours", "2h", new(time.Duration), 2 * time.Hour, false},
		{"duration milliseconds", "500ms", new(time.Duration), 500 * time.Millisecond, false},
		{"duration microseconds", "100us", new(time.Duration), 100 * time.Microsecond, false},
		{"duration nanoseconds", "1000ns", new(time.Duration), 1000 * time.Nanosecond, false},
		{"duration combined standard", "1h30m45s", new(time.Duration), 1*time.Hour + 30*time.Minute + 45*time.Second, false},

		// Duration - day suffix 'd'
		{"duration days", "5d", new(time.Duration), 5 * 24 * time.Hour, false},
		{"duration one day", "1d", new(time.Duration), 24 * time.Hour, false},
		{"duration uppercase D", "3D", new(time.Duration), 3 * 24 * time.Hour, false},
		{"duration zero days", "0d", new(time.Duration), time.Duration(0), false},
		{"duration fractional days half", "0.5d", new(time.Duration), 12 * time.Hour, false},
		{"duration fractional days quarter", "0.25d", new(time.Duration), 6 * time.Hour, false},
		{"duration fractional days third", "1.5d", new(time.Duration), 36 * time.Hour, false},

		// Duration - days combined with other units
		{"duration days and hours", "1d12h", new(time.Duration), 36 * time.Hour, false},
		{"duration days hours minutes", "2d3h30m", new(time.Duration), 51*time.Hour + 30*time.Minute, false},
		{"duration days hours minutes seconds", "1d2h3m4s", new(time.Duration), 26*time.Hour + 3*time.Minute + 4*time.Second, false},
		{"duration days and seconds only", "1d30s", new(time.Duration), 24*time.Hour + 30*time.Second, false},
		{"duration days and milliseconds", "1d500ms", new(time.Duration), 24*time.Hour + 500*time.Millisecond, false},

		// Duration - negative values
		{"duration negative days", "-1d", new(time.Duration), -24 * time.Hour, false},
		{"duration negative combined", "-2d12h", new(time.Duration), -60 * time.Hour, false},

		// Duration - large values
		{"duration week equivalent", "7d", new(time.Duration), 7 * 24 * time.Hour, false},
		{"duration month equivalent", "30d", new(time.Duration), 30 * 24 * time.Hour, false},
		{"duration year equivalent", "365d", new(time.Duration), 365 * 24 * time.Hour, false},

		// Duration - error cases
		{"duration invalid format", "5x", new(time.Duration), nil, true},
		{"duration empty", "", new(time.Duration), nil, true},
		{"duration only letters", "abc", new(time.Duration), nil, true},

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
