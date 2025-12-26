package tags

import (
	"reflect"

	"github.com/arloliu/fuda/internal/types"
)

// ProcessDefault processes the 'default' tag for a field.
func ProcessDefault(field reflect.StructField, value reflect.Value) error {
	tag := field.Tag.Get("default")
	if tag == "" || tag == "-" {
		return nil
	}

	// Only set default if value is zero
	if !value.IsZero() {
		return nil
	}

	return types.Convert(tag, value)
}
