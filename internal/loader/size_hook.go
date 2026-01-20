package loader

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/arloliu/fuda/internal/bytesize"
	"gopkg.in/yaml.v3"
)

var sizePattern = regexp.MustCompile(`^([\d.]+)([a-zA-Z]+)$`)

// preprocessSizeNodesForType walks a YAML node tree and converts size string values
// to integer values (bytes), but only for numeric target fields.
// Examples: "1KiB" → "1024", "1GB" → "1000000000"
//
// This avoids coercing values for string fields while still supporting size strings
// for numeric fields in structs.
func preprocessSizeNodesForType(node *yaml.Node, targetType reflect.Type) {
	if node == nil {
		return
	}
	if targetType != nil && targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			preprocessSizeNodesForType(child, targetType)
		}
	case yaml.MappingNode:
		switch {
		case targetType != nil && targetType.Kind() == reflect.Struct:
			fieldMap := yamlFieldTypeMap(targetType)
			for i := 0; i < len(node.Content); i += 2 {
				keyNode := node.Content[i]
				valNode := node.Content[i+1]
				if keyNode.Kind != yaml.ScalarNode {
					continue
				}
				fieldType, ok := fieldMap[keyNode.Value]
				if !ok {
					continue
				}
				preprocessSizeNodesForType(valNode, fieldType)
			}
		case targetType != nil && targetType.Kind() == reflect.Map:
			valType := targetType.Elem()
			for i := 0; i < len(node.Content); i += 2 {
				preprocessSizeNodesForType(node.Content[i+1], valType)
			}
		default:
			// Unknown target type; avoid coercion
		}
	case yaml.ScalarNode:
		// Only process string nodes that look like size strings and map to numeric types
		if node.Tag == "!!str" && isNumericType(targetType) {
			if matches := sizePattern.FindStringSubmatch(node.Value); len(matches) == 3 {
				numStr := matches[1]
				unitStr := matches[2]

				if val, ok := bytesize.ParseToBigInt(numStr, unitStr); ok {
					// Update node to be an integer
					node.Tag = "!!int"
					node.Value = val.String()
				}
			}
		}
	case yaml.AliasNode:
		// Aliases are resolved by yaml.Decode, no preprocessing needed
	}
}

func isNumericType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	//nolint:exhaustive // Only numeric kinds should return true.
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func yamlFieldTypeMap(t reflect.Type) map[string]reflect.Type {
	result := make(map[string]reflect.Type)
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		addTag := func(tagKey string) {
			tag := field.Tag.Get(tagKey)
			name := strings.Split(tag, ",")[0]
			if name == "-" {
				return
			}
			if name == "" {
				name = field.Name
			}
			result[name] = field.Type
		}

		addTag("yaml")
		addTag("json")
	}

	return result
}
