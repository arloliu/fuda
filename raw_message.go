package fuda

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// RawMessage stores raw configuration data for deferred unmarshaling.
// Use this for polymorphic config where the struct type depends on a discriminator field.
//
// Example:
//
//	type DeviceConfig struct {
//	    Type       string          `yaml:"type"`
//	    Properties fuda.RawMessage `yaml:"properties"`
//	}
//
//	// After loading, unmarshal Properties based on Type:
//	switch cfg.Type {
//	case "car":
//	    var props CarProperties
//	    cfg.Properties.Unmarshal(&props)
//	}
type RawMessage []byte

// MarshalJSON returns m as-is. User is responsible for ensuring valid JSON.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON stores a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("fuda.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (m RawMessage) MarshalYAML() (any, error) {
	if m == nil {
		return nil, nil //nolint:nilnil // nil is valid YAML null
	}
	var node yaml.Node
	if err := yaml.Unmarshal(m, &node); err != nil {
		return nil, err
	}
	// yaml.Unmarshal creates a Document node; return its content
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0], nil
	}

	return &node, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (m *RawMessage) UnmarshalYAML(node *yaml.Node) error {
	if m == nil {
		return errors.New("fuda.RawMessage: UnmarshalYAML on nil pointer")
	}
	out, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	*m = append((*m)[0:0], out...)

	return nil
}

// Unmarshal decodes the raw message into v.
// Uses YAML unmarshaler which handles both YAML and JSON.
func (m RawMessage) Unmarshal(v any) error {
	return yaml.Unmarshal(m, v)
}
