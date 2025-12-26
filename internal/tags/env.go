package tags

import (
	"os"
	"reflect"

	"github.com/arloliu/fuda/internal/types"
)

// ProcessEnv processes the 'env' tag for a field.
func ProcessEnv(field reflect.StructField, value reflect.Value, prefix string) error {
	tag := field.Tag.Get("env")
	if tag == "" {
		return nil
	}

	envKey := tag
	if prefix != "" {
		envKey = prefix + envKey
	}

	envVal, ok := os.LookupEnv(envKey)
	if !ok {
		return nil
	}

	return types.Convert(envVal, value)
}
