// Package fuda provides struct-tag-first configuration loading for Go.
//
// It supports loading configuration from YAML/JSON files, environment variables,
// and external references (files, HTTP endpoints), with built-in defaults and validation.
//
// Basic usage:
//
//	type Config struct {
//	    Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
//	    Port int    `yaml:"port" default:"8080"`
//	}
//
//	var cfg Config
//	if err := fuda.LoadFile("config.yaml", &cfg); err != nil {
//	    log.Fatal(err)
//	}
//
// For more control, use the Builder pattern:
//
//	loader, err := fuda.New().
//	    FromFile("config.yaml").
//	    WithEnvPrefix("APP_").
//	    Build()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := loader.Load(&cfg); err != nil {
//	    log.Fatal(err)
//	}
package fuda

import (
	"io"
	"os"
	"reflect"
	"text/template"
	"time"

	"github.com/arloliu/fuda/internal/loader"
	"github.com/arloliu/fuda/internal/resolver"
	"github.com/go-playground/validator/v10"
)

// Loader is responsible for loading configuration from various sources.
type Loader struct {
	loaderConfig
	source     []byte
	sourceName string
}

// loaderConfig holds the configuration for the loader.
type loaderConfig struct {
	envPrefix   string
	validator   *validator.Validate
	refResolver RefResolver
	timeout     time.Duration
	tmplConfig  *templateConfig
	tmplData    any
}

// templateConfig holds template parsing configuration.
type templateConfig struct {
	leftDelim  string
	rightDelim string
	missingKey string
	funcMap    template.FuncMap
}

// TemplateOption configures template parsing behavior.
type TemplateOption func(*templateConfig)

// WithDelimiters sets custom delimiters for template parsing.
// Use this when your configuration contains literal "{{" sequences that
// should not be interpreted as template syntax.
//
// Example:
//
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithTemplate(data, fuda.WithDelimiters("<{", "}>")).
//	    Build()
func WithDelimiters(left, right string) TemplateOption {
	return func(c *templateConfig) {
		c.leftDelim = left
		c.rightDelim = right
	}
}

// WithMissingKey controls behavior when a map is indexed with a key not in the map.
// Valid values:
//   - "invalid" (default): no error, outputs "<no value>"
//   - "zero": outputs the zero value for the type
//   - "error": template execution stops with an error
func WithMissingKey(behavior string) TemplateOption {
	return func(c *templateConfig) {
		c.missingKey = behavior
	}
}

// WithFuncs adds custom template functions.
// These are merged with the template's built-in functions.
func WithFuncs(funcMap template.FuncMap) TemplateOption {
	return func(c *templateConfig) {
		c.funcMap = funcMap
	}
}

// New creates a new configuration Builder.
func New() *Builder {
	return &Builder{
		config: loaderConfig{
			validator: validator.New(),
		},
	}
}

// Builder provides a fluent API for constructing a Loader.
type Builder struct {
	config loaderConfig
	source []byte
	name   string
	err    error
}

// FromFile reads configuration from the file at path.
// The file format (YAML or JSON) is auto-detected from content.
func (b *Builder) FromFile(path string) *Builder {
	if b.err != nil {
		return b
	}

	data, err := os.ReadFile(path)
	if err != nil {
		b.err = err

		return b
	}

	b.source = data
	b.name = path

	return b
}

// FromReader reads configuration from an io.Reader.
// The content format (YAML or JSON) is auto-detected.
func (b *Builder) FromReader(r io.Reader) *Builder {
	if b.err != nil {
		return b
	}

	data, err := io.ReadAll(r)
	if err != nil {
		b.err = err

		return b
	}

	b.source = data
	b.name = "reader"

	return b
}

// FromBytes uses the provided byte slice as configuration data.
// The content format (YAML or JSON) is auto-detected.
func (b *Builder) FromBytes(data []byte) *Builder {
	b.source = data
	b.name = "bytes"

	return b
}

// WithEnvPrefix sets a prefix for environment variable lookups.
// For example, with prefix "APP_", an `env:"HOST"` tag reads APP_HOST.
func (b *Builder) WithEnvPrefix(prefix string) *Builder {
	b.config.envPrefix = prefix

	return b
}

// WithValidator sets a custom validator instance.
// If not set, a default validator is used.
func (b *Builder) WithValidator(v *validator.Validate) *Builder {
	b.config.validator = v

	return b
}

// WithRefResolver sets a custom reference resolver for ref/refFrom tags.
// The default resolver supports file://, http://, and https:// schemes.
func (b *Builder) WithRefResolver(r RefResolver) *Builder {
	b.config.refResolver = r

	return b
}

// WithTimeout sets a timeout for reference resolution (ref/refFrom tags).
// Default is 0 (no timeout). Set explicitly for network refs.
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.config.timeout = timeout

	return b
}

// Apply applies a configuration function to the builder.
// This enables reusable configuration bundles:
//
//	var prodConfig = func(b *fuda.Builder) {
//	    b.WithEnvPrefix("PROD_").WithTimeout(10 * time.Second)
//	}
//	loader, _ := fuda.New().FromFile("config.yaml").Apply(prodConfig).Build()
func (b *Builder) Apply(fn func(*Builder)) *Builder {
	fn(b)

	return b
}

// WithTemplate enables Go template processing on configuration content before YAML parsing.
// The data parameter provides template context, and opts configure template behavior.
//
// Template processing occurs BEFORE YAML unmarshaling. If your configuration contains
// literal "{{" or "}}" sequences that should not be interpreted as template delimiters,
// use WithDelimiters to specify alternative delimiters.
//
// Example:
//
//	type TemplateData struct {
//	    Host string
//	    Port int
//	}
//
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithTemplate(TemplateData{Host: "localhost", Port: 8080}).
//	    Build()
func (b *Builder) WithTemplate(data any, opts ...TemplateOption) *Builder {
	cfg := &templateConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	b.config.tmplConfig = cfg
	b.config.tmplData = data

	return b
}

// Build creates the Loader with the configured options.
// Returns an error if any prior builder method (FromFile, FromReader) failed.
func (b *Builder) Build() (*Loader, error) {
	if b.err != nil {
		return nil, b.err
	}

	// Use default resolver if not provided
	refResolver := b.config.refResolver
	if refResolver == nil {
		refResolver = resolver.New()
	}

	return &Loader{
		loaderConfig: loaderConfig{
			envPrefix:   b.config.envPrefix,
			validator:   b.config.validator,
			refResolver: refResolver,
			timeout:     b.config.timeout,
			tmplConfig:  b.config.tmplConfig,
			tmplData:    b.config.tmplData,
		},
		source:     b.source,
		sourceName: b.name,
	}, nil
}

// Load populates the target struct with configuration.
func (l *Loader) Load(target any) error {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Pointer || targetVal.IsNil() {
		return &FieldError{Message: "target must be a non-nil pointer"}
	}

	var tmplCfg *loader.TemplateConfig
	if l.tmplConfig != nil {
		tmplCfg = &loader.TemplateConfig{
			LeftDelim:  l.tmplConfig.leftDelim,
			RightDelim: l.tmplConfig.rightDelim,
			MissingKey: l.tmplConfig.missingKey,
			FuncMap:    l.tmplConfig.funcMap,
		}
	}

	engine := &loader.Engine{
		Validator:      l.validator,
		RefResolver:    l.refResolver,
		EnvPrefix:      l.envPrefix,
		Source:         l.source,
		SourceName:     l.sourceName,
		Timeout:        l.timeout,
		TemplateConfig: tmplCfg,
		TemplateData:   l.tmplData,
	}

	return engine.Load(target)
}

// Convenience Functions

// LoadFile parses the file at path and populates target.
func LoadFile(path string, target any) error {
	l, err := New().FromFile(path).Build()
	if err != nil {
		return err
	}

	return l.Load(target)
}

// LoadBytes parses the raw bytes and populates target.
func LoadBytes(data []byte, target any) error {
	l, err := New().FromBytes(data).Build()
	if err != nil {
		return err
	}

	return l.Load(target)
}

// LoadReader reads from r and populates target.
func LoadReader(r io.Reader, target any) error {
	l, err := New().FromReader(r).Build()
	if err != nil {
		return err
	}

	return l.Load(target)
}
