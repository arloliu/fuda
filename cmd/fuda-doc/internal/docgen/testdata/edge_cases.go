package testdata

// Flat is a struct with only scalar fields and no nesting.
type Flat struct {
	// Name is the display name.
	Name string `yaml:"name" default:"flat" env:"FLAT_NAME"`

	// Count is an integer counter.
	Count int `yaml:"count" default:"10"`

	// Enabled toggles the feature.
	Enabled bool `yaml:"enabled" default:"true" env:"FLAT_ENABLED"`
}

// WithPointer has a pointer to another same-package struct.
type WithPointer struct {
	// Label identifies this item.
	Label string `yaml:"label"`

	// Inner is an optional nested config.
	Inner *InnerConfig `yaml:"inner,omitempty"`
}

// InnerConfig is used by WithPointer.
type InnerConfig struct {
	// Value holds the inner value.
	Value string `yaml:"value" default:"inner-default" env:"INNER_VALUE"`

	// Retries is the retry count.
	Retries int `yaml:"retries" default:"3"`
}

// WithSliceAndMap has slice and map fields.
type WithSliceAndMap struct {
	// Items is a list of strings.
	Items []string `yaml:"items" default:"a,b,c"`

	// Labels is a map of key-value pairs.
	Labels map[string]string `yaml:"labels" default:"env:dev,tier:web"`

	// Flags is a map of feature flags.
	Flags map[string]bool `yaml:"flags"`
}

// WithAllTags exercises every supported tag.
type WithAllTags struct {
	// FieldA has all the tags.
	FieldA string `yaml:"field_a" default:"hello" env:"FIELD_A" validate:"required" ref:"file:///tmp/a" json:"fieldA"`

	// FieldB uses refFrom.
	FieldB string `yaml:"field_b,omitempty" refFrom:"FieldBPath"`

	// FieldBPath is the path for FieldB.
	FieldBPath string `yaml:"field_b_path" default:"/secrets/b" env:"FIELD_B_PATH"`

	// FieldC uses dsn tag.
	FieldC string `yaml:"field_c" dsn:"{{.User}}:{{.Pass}}@{{.Host}}"`
}

// DeepNest has three levels of same-package nesting.
type DeepNest struct {
	// Level1 is the first nesting level.
	Level1 Level1Config `yaml:"level1"`
}

// Level1Config is first level.
type Level1Config struct {
	// Name at level 1.
	Name string `yaml:"name" default:"l1"`

	// Level2 nests deeper.
	Level2 Level2Config `yaml:"level2"`
}

// Level2Config is second level.
type Level2Config struct {
	// Name at level 2.
	Name string `yaml:"name" default:"l2"`

	// Level3 nests even deeper.
	Level3 Level3Config `yaml:"level3"`
}

// Level3Config is the deepest level.
type Level3Config struct {
	// Value at the bottom.
	Value string `yaml:"value" default:"deep"`
}

// WithEmbedded uses an embedded struct (no field name).
type WithEmbedded struct {
	// Visible is a regular field.
	Visible string `yaml:"visible" default:"yes"`

	// EmbeddedMeta is embedded.
	EmbeddedMeta
}

// EmbeddedMeta is meant to be embedded.
type EmbeddedMeta struct {
	// Version of the config.
	Version string `yaml:"version" default:"v1"`

	// Author of the config.
	Author string `yaml:"author" default:"system"`
}

// NoTags has fields without any struct tags.
type NoTags struct {
	// FieldX has no tags at all.
	FieldX string

	// FieldY also has no tags.
	FieldY int
}

// NoComments has fields without any doc comments.
type NoComments struct {
	Alpha string `yaml:"alpha" default:"a"`
	Beta  int    `yaml:"beta" default:"1"`
}
