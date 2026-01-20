package loader

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPreprocessSizeNodes(t *testing.T) {
	type intConfig struct {
		Val int `yaml:"val"`
	}
	type stringConfig struct {
		Val string `yaml:"val"`
	}
	type jsonConfig struct {
		Val int `json:"val"`
	}

	tests := []struct {
		name     string
		input    string
		expected string // expected tag
		wantVal  string
		typeOf   reflect.Type
	}{
		{
			name:     "simple kib",
			input:    "val: 1KiB",
			expected: "!!int",
			wantVal:  "1024",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "simple kb (SI)",
			input:    "val: 1KB",
			expected: "!!int",
			wantVal:  "1000",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "decimal value",
			input:    "val: 0.5MiB",
			expected: "!!int",
			wantVal:  "524288",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "large value GB",
			input:    "val: 2GB",
			expected: "!!int",
			wantVal:  "2000000000",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "unquoted string in yaml is usually scalar",
			input:    "val: 1KiB",
			expected: "!!int",
			wantVal:  "1024",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "ignore pure string",
			input:    "val: someString",
			expected: "!!str",
			wantVal:  "someString",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "ignore raw number string (let yaml handle or type converter)",
			input:    `val: "123"`, // Quoted number
			expected: "!!str",
			wantVal:  "123",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "case insensitive match",
			input:    "val: 1kib",
			expected: "!!int",
			wantVal:  "1024",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "quoted size string is also converted",
			input:    `val: "1KiB"`,
			expected: "!!int",
			wantVal:  "1024",
			typeOf:   reflect.TypeFor[intConfig](),
		},
		{
			name:     "string field does not coerce",
			input:    "val: 1KiB",
			expected: "!!str",
			wantVal:  "1KiB",
			typeOf:   reflect.TypeFor[stringConfig](),
		},
		{
			name:     "json tag field coerces",
			input:    "val: 1KiB",
			expected: "!!int",
			wantVal:  "1024",
			typeOf:   reflect.TypeFor[jsonConfig](),
		},
		{
			name:     "json input with json tag",
			input:    `{"val":"1KiB"}`,
			expected: "!!int",
			wantVal:  "1024",
			typeOf:   reflect.TypeFor[jsonConfig](),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(tt.input), &node))

			preprocessSizeNodesForType(&node, tt.typeOf)

			// The structure of node for "key: value" is:
			// Document -> Mapping -> [KeyNode, ValueNode]
			// We check the ValueNode
			require.NotEmpty(t, node.Content)
			require.Len(t, node.Content[0].Content, 2)
			valNode := node.Content[0].Content[1]
			require.Equal(t, tt.expected, valNode.Tag)
			require.Equal(t, tt.wantVal, valNode.Value)
		})
	}
}
