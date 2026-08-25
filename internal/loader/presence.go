package loader

import (
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// This file tracks which fields the YAML document actually supplied,
// so the default tag can fill a field only when nothing else did.
// The lookup rules deliberately mirror yaml.v3's decoder:
// keys match the yaml tag name or the lowercased field name (case-sensitively),
// embedded structs nest under their type name unless tagged inline,
// aliases follow their anchor, and merge keys ("<<") pull in the merged mappings.
// A key only matches when, after resolving aliases, it is a scalar node
// whose resolved tag is neither "!!null" nor "!!binary" and whose raw
// text equals the name.
// yaml.v3's decoder turns any scalar key into a struct field name or map
// key by taking its literal text (Node.Value) regardless of the key's
// resolved type, so a plain "true:" or "0:" key decodes exactly like a
// quoted "true:" or "0:" key would.
// The one key shape the decoder actually refuses is an implicit null
// (an unquoted null, Null, NULL, or ~, verified empirically for all four):
// decoding null into a string always fails, so the entry never reaches
// any field or map key, not even one literally named "null".
// A quoted "null" is unaffected, since quoting forces the string tag
// rather than resolving to null.
// A "!!binary" key is the other exception, and it is not refused, only
// decoded differently: the decoder base64-decodes the key's text and
// uses the DECODED bytes as the real key, not the raw base64 text.
// Matching a binary key by its raw text would therefore be wrong,
// so this lookup never matches one, and it always reports absent for it
// instead of duplicating the decoder's base64 handling.
// That is a safe degradation, not a gap: a binary key can only ever
// cause a false negative here (default applied when it need not be),
// never the false positive that raw-text matching risked.
// When a shape cannot be resolved the lookup reports the field as absent,
// which preserves the previous behavior of applying the default.

// maxYAMLIndirection bounds alias and merge resolution so a pathological
// document cannot spin the lookup.
const maxYAMLIndirection = 32

var (
	yamlUnmarshalerType         = reflect.TypeOf((*yaml.Unmarshaler)(nil)).Elem()
	yamlObsoleteUnmarshalerType = reflect.TypeOf((*yamlObsoleteUnmarshaler)(nil)).Elem()
)

// yamlObsoleteUnmarshaler is yaml.v3's legacy custom-unmarshal interface.
type yamlObsoleteUnmarshaler interface {
	UnmarshalYAML(unmarshal func(any) error) error
}

// resolveYAMLNode unwraps document and alias indirection until it
// reaches a concrete node.
func resolveYAMLNode(node *yaml.Node) *yaml.Node {
	for range maxYAMLIndirection {
		switch {
		case node == nil:
			return nil
		case node.Kind == yaml.DocumentNode && len(node.Content) > 0:
			node = node.Content[0]
		case node.Kind == yaml.AliasNode:
			node = node.Alias
		default:
			return node
		}
	}

	return nil
}

// yamlNodeSupplied reports whether the node represents a value the document actually supplied.
// An explicit null is treated as absent:
// yaml.v3 ignores null when decoding,
// so the field keeps its zero value and the default should still apply.
func yamlNodeSupplied(node *yaml.Node) bool {
	return node != nil && (node.Kind != yaml.ScalarNode || node.Tag != "!!null")
}

// yamlFieldNode returns the value node the document supplied for the struct field, if any.
// For a field tagged yaml:",inline" it returns the mapping itself with inline=true,
// because the field's own keys live directly in the parent mapping.
func yamlFieldNode(node *yaml.Node, field reflect.StructField) (child *yaml.Node, inline bool) {
	mapping := resolveYAMLNode(node)
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil, false
	}

	name, isInline := yamlFieldKey(field)
	if isInline {
		return mapping, true
	}
	if name == "" {
		return nil, false
	}

	return lookupYAMLKey(mapping, name, 0), false
}

// yamlFieldKey returns the document key for a struct field,
// following yaml.v3's rules:
// the yaml tag name when present, otherwise the lowercased field name.
// An empty name means the field is not addressable from the document (yaml:"-").
func yamlFieldKey(field reflect.StructField) (name string, inline bool) {
	parts := strings.Split(field.Tag.Get("yaml"), ",")
	name = parts[0]
	for _, opt := range parts[1:] {
		if opt == "inline" {
			return "", true
		}
	}

	if name == "-" {
		return "", false
	}
	if name == "" {
		name = strings.ToLower(field.Name)
	}

	return name, false
}

// lookupYAMLKey finds the value node for a key in a mapping,
// resolving merge keys ("<<") the way yaml.v3's decoder does.
// Keys stated directly in the mapping take precedence over merged ones.
func lookupYAMLKey(mapping *yaml.Node, name string, depth int) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode || depth > maxYAMLIndirection {
		return nil
	}

	var merges []*yaml.Node
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key, val := mapping.Content[i], mapping.Content[i+1]
		if key.Tag == "!!merge" {
			merges = append(merges, val)
			continue
		}
		if isStringKeyMatch(key, name) {
			return resolveYAMLNode(val)
		}
	}

	// A merge key's value is a mapping, an alias to one, or a sequence
	// of those.
	for _, merge := range merges {
		merge = resolveYAMLNode(merge)
		if merge == nil {
			continue
		}
		if merge.Kind == yaml.SequenceNode {
			for _, item := range merge.Content {
				if found := lookupYAMLKey(resolveYAMLNode(item), name, depth+1); found != nil {
					return found
				}
			}

			continue
		}
		if found := lookupYAMLKey(merge, name, depth+1); found != nil {
			return found
		}
	}

	return nil
}

// isStringKeyMatch reports whether a mapping key node decodes to the
// plain string value name, the way yaml.v3's decoder would decode it
// into a struct field name or a string map key.
// The decoder takes the key's literal text regardless of its resolved
// type (bool, int, float, and plain string keys all match by raw text),
// with two exceptions.
// A key that resolves to null never decodes into any name,
// so it is excluded here.
// A "!!binary" key is base64-decoded by the decoder, and the decoded
// bytes become the string key, not the raw base64 text,
// so raw-text matching would be wrong for it too.
// Rather than duplicate the decoder's base64 handling here,
// a binary key is always reported as not matching,
// which is the safe direction: it degrades to "absent",
// so the default still applies instead of a false "supplied".
func isStringKeyMatch(key *yaml.Node, name string) bool {
	resolved := resolveYAMLNode(key)
	if resolved == nil || resolved.Kind != yaml.ScalarNode {
		return false
	}
	if resolved.Tag == "!!null" || resolved.Tag == "!!binary" {
		return false
	}

	return resolved.Value == name
}

// yamlSequenceElem returns the node for element i of a sequence,
// or nil when the node is not a sequence or the index is out of range.
func yamlSequenceElem(node *yaml.Node, i int) *yaml.Node {
	seq := resolveYAMLNode(node)
	if seq == nil || seq.Kind != yaml.SequenceNode || i >= len(seq.Content) {
		return nil
	}

	return resolveYAMLNode(seq.Content[i])
}

// yamlMapValue returns the node for a map entry.
// Only string keys are resolved;
// other key types report absent, which keeps the previous default behavior for those maps.
func yamlMapValue(node *yaml.Node, key reflect.Value) *yaml.Node {
	mapping := resolveYAMLNode(node)
	if mapping == nil || mapping.Kind != yaml.MappingNode || key.Kind() != reflect.String {
		return nil
	}

	return lookupYAMLKey(mapping, key.String(), 0)
}

// hasCustomYAMLUnmarshaler reports whether the type decodes itself.
// The document keys of such a type need not correspond to its Go fields,
// so presence tracking stops there and fields keep the previous default behavior.
func hasCustomYAMLUnmarshaler(t reflect.Type) bool {
	ptr := reflect.PointerTo(t)

	return ptr.Implements(yamlUnmarshalerType) || ptr.Implements(yamlObsoleteUnmarshalerType)
}
