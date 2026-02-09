package tags

import (
	"os"
	"reflect"

	"github.com/arloliu/fuda/internal/types"
)

// ProcessEnv processes the 'env' tag for a field.
// Returns true if an environment variable was found and applied, false otherwise.
// Environment variables always override current values when the env var is set.
func ProcessEnv(field reflect.StructField, value reflect.Value, prefix string) (bool, error) {
	tag := field.Tag.Get("env")
	if tag == "" {
		return false, nil
	}

	envKey := tag
	if prefix != "" {
		envKey = prefix + envKey
	}

	envVal, ok := os.LookupEnv(envKey)
	if !ok {
		return false, nil
	}

	return true, types.Convert(envVal, value)
}
