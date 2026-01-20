package loader

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPreprocessDurationNodes_JSON(t *testing.T) {
	input := `{"timeout":"2d","nested":{"delay":"1d12h"}}`

	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(input), &node))

	preprocessDurationNodes(&node)

	timeoutNode := findMappingValue(&node, "timeout")
	require.NotNil(t, timeoutNode)
	require.Equal(t, "48h", timeoutNode.Value)

	nestedNode := findMappingValue(&node, "nested")
	require.NotNil(t, nestedNode)
	delayNode := findMappingValue(nestedNode, "delay")
	require.NotNil(t, delayNode)
	require.Equal(t, "24h12h", delayNode.Value)
}

func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil {
		return nil
	}

	// Descend to mapping node if needed
	//nolint:exhaustive // Only handle relevant node kinds for test helper.
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil
		}
		node = node.Content[0]
	case yaml.MappingNode:
		// use as is
	default:
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]
		if keyNode.Kind == yaml.ScalarNode && keyNode.Value == key {
			return valNode
		}
	}

	return nil
}
