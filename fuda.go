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
	"bytes"
	"io"
	"reflect"
	"text/template"
	"time"

	"github.com/arloliu/fuda/internal/loader"
	"github.com/arloliu/fuda/internal/resolver"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/yaml/kyaml"
)

// Loader is responsible for loading configuration from various sources.
type Loader struct {
	loaderConfig
	source     []byte
	sourceName string
}

// loaderConfig holds the configuration for the loader.
type loaderConfig struct {
	fs           afero.Fs // Filesystem for file operations
	envPrefix    string
	validator    *validator.Validate
	refResolver  RefResolver
	timeout      time.Duration
	tmplConfig   *templateConfig
	tmplData     any
	dotenvConfig *dotenvConfig  // dotenv file loading configuration
	overrides    map[string]any // Programmatic value overrides
	// Preprocessing toggles (nil means default true)
	enableSizePreprocess     *bool
	enableDurationPreprocess *bool
}

// dotenvConfig holds dotenv file loading configuration.
type dotenvConfig struct {
	files       []string // Explicit file paths to load
	searchPaths []string // Directories to search for env file
	searchName  string   // Filename to search for (e.g., ".env")
	override    bool     // If true, use godotenv.Overload instead of Load
}

// DotEnvOption configures dotenv loading behavior.
type DotEnvOption func(*dotenvConfig)

// DotEnvOverride returns an option that causes dotenv values to override
// existing environment variables. By default, existing env vars take precedence.
func DotEnvOverride() DotEnvOption {
	return func(c *dotenvConfig) {
		c.override = true
	}
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

	fs := b.config.fs
	if fs == nil {
		fs = DefaultFs
	}

	data, err := afero.ReadFile(fs, path)
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

// WithFilesystem sets a custom filesystem for file operations.
// This is useful for testing with in-memory filesystems.
//
// Example:
//
//	memFs := afero.NewMemMapFs()
//	afero.WriteFile(memFs, "/config.yaml", []byte("host: localhost"), 0644)
//	loader, _ := fuda.New().
//	    WithFilesystem(memFs).
//	    FromFile("/config.yaml").
//	    Build()
func (b *Builder) WithFilesystem(fs afero.Fs) *Builder {
	b.config.fs = fs

	return b
}

// WithTimeout sets a timeout for reference resolution (ref/refFrom tags).
// Default is 0 (no timeout). Set explicitly for network refs.
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.config.timeout = timeout

	return b
}

// WithOverrides sets programmatic overrides that take precedence over config file values.
// These are applied after template processing but before struct unmarshaling.
// Keys use dot notation for nested values: "database.host" overrides database.host.
//
// Example:
//
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithOverrides(map[string]any{
//	        "host": "override.example.com",
//	        "database.port": 5433,
//	    }).
//	    Build()
func (b *Builder) WithOverrides(overrides map[string]any) *Builder {
	b.config.overrides = overrides

	return b
}

// WithSizePreprocess enables or disables size-string preprocessing.
// Default is enabled for backward compatibility.
func (b *Builder) WithSizePreprocess(enabled bool) *Builder {
	b.config.enableSizePreprocess = &enabled

	return b
}

// WithDurationPreprocess enables or disables duration-string preprocessing.
// Default is enabled for backward compatibility.
func (b *Builder) WithDurationPreprocess(enabled bool) *Builder {
	b.config.enableDurationPreprocess = &enabled

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

// WithDotEnv loads environment variables from a dotenv file before processing.
// The file is loaded before any `env` tag resolution, so dotenv values become
// available to struct fields with env tags.
//
// By default, existing environment variables take precedence over dotenv values.
// Use DotEnvOverride() option to reverse this behavior.
//
// Missing files are silently ignored, making this safe for optional .env.local files.
//
// Example:
//
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithDotEnv(".env").
//	    Build()
//
//	// With override mode:
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithDotEnv(".env", fuda.DotEnvOverride()).
//	    Build()
func (b *Builder) WithDotEnv(file string, opts ...DotEnvOption) *Builder {
	cfg := &dotenvConfig{
		files: []string{file},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	b.config.dotenvConfig = cfg

	return b
}

// WithDotEnvFiles loads environment variables from multiple dotenv files.
// Files are loaded in order; later files can supplement earlier ones.
// This enables environment-specific overlays:
//
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithDotEnvFiles([]string{".env", ".env.local", ".env.production"}).
//	    Build()
//
// Missing files are silently ignored.
// Use DotEnvOverride() option to override existing env vars.
func (b *Builder) WithDotEnvFiles(files []string, opts ...DotEnvOption) *Builder {
	cfg := &dotenvConfig{
		files: files,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	b.config.dotenvConfig = cfg

	return b
}

// WithDotEnvSearch searches for a dotenv file in the specified directories.
// The first existing file found wins. This is useful when the application
// may be run from different working directories:
//
//	loader, _ := fuda.New().
//	    FromFile("config.yaml").
//	    WithDotEnvSearch(".env", []string{".", "./config", "/etc/myapp"}).
//	    Build()
//
// The name parameter is the filename to search for (e.g., ".env").
// Use DotEnvOverride() option to override existing env vars.
func (b *Builder) WithDotEnvSearch(name string, searchPaths []string, opts ...DotEnvOption) *Builder {
	cfg := &dotenvConfig{
		searchName:  name,
		searchPaths: searchPaths,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	b.config.dotenvConfig = cfg

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
		fs := b.config.fs
		if fs == nil {
			fs = DefaultFs
		}
		refResolver = resolver.New(fs)
	}

	return &Loader{
		loaderConfig: loaderConfig{
			envPrefix:                b.config.envPrefix,
			validator:                b.config.validator,
			refResolver:              refResolver,
			timeout:                  b.config.timeout,
			tmplConfig:               b.config.tmplConfig,
			tmplData:                 b.config.tmplData,
			dotenvConfig:             b.config.dotenvConfig,
			overrides:                b.config.overrides,
			enableSizePreprocess:     b.config.enableSizePreprocess,
			enableDurationPreprocess: b.config.enableDurationPreprocess,
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

	var dotenvCfg *loader.DotenvConfig
	if l.dotenvConfig != nil {
		dotenvCfg = &loader.DotenvConfig{
			Files:       l.dotenvConfig.files,
			SearchPaths: l.dotenvConfig.searchPaths,
			SearchName:  l.dotenvConfig.searchName,
			Override:    l.dotenvConfig.override,
		}
	}

	engine := &loader.Engine{
		Validator:                l.validator,
		RefResolver:              l.refResolver,
		EnvPrefix:                l.envPrefix,
		Source:                   l.source,
		SourceName:               l.sourceName,
		Timeout:                  l.timeout,
		TemplateConfig:           tmplCfg,
		TemplateData:             l.tmplData,
		DotenvConfig:             dotenvCfg,
		Overrides:                l.overrides,
		EnableSizePreprocess:     l.enableSizePreprocess,
		EnableDurationPreprocess: l.enableDurationPreprocess,
	}

	return engine.Load(target)
}

// ToKYAML converts the loader's source to KYAML format.
// Returns an error if the source is not valid YAML format.
// KYAML is a strict subset of YAML that is explicit and unambiguous,
// designed to be halfway between YAML and JSON.
func (l *Loader) ToKYAML() ([]byte, error) {
	if len(l.source) == 0 {
		return nil, &FieldError{Message: "no source data to convert"}
	}

	// Validate that the source is valid YAML by unmarshaling to a generic type
	var test any
	if err := yaml.Unmarshal(l.source, &test); err != nil {
		return nil, &FieldError{Message: "source is not valid YAML format", Err: err}
	}

	// Convert to KYAML format
	var buf bytes.Buffer
	encoder := &kyaml.Encoder{}
	if err := encoder.FromYAML(bytes.NewReader(l.source), &buf); err != nil {
		return nil, &FieldError{Message: "failed to convert to KYAML format", Err: err}
	}

	return buf.Bytes(), nil
}

// ToMap returns the loader's source configuration as a map.
// This returns the raw source data (before struct loading, defaults, or env overrides).
// Useful for debugging, logging, or passing configuration to other systems.
// Returns an error if no source is set or if YAML parsing fails.
func (l *Loader) ToMap() (map[string]any, error) {
	if len(l.source) == 0 {
		return nil, &FieldError{Message: "no source data to convert"}
	}

	var result map[string]any
	if err := yaml.Unmarshal(l.source, &result); err != nil {
		return nil, &FieldError{Message: "source is not valid YAML/JSON", Err: err}
	}

	return result, nil
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

// SetDefaults applies `default` tag values to the target struct.
// Environment variables (via `env` tag) and references (`ref`, `refFrom`) are also resolved,
// but no YAML/JSON source is processed.
//
// By default, NO validation is performed. Use WithValidation(true) to enable validation.
//
// Example:
//
//	type Config struct {
//	    Host    string `default:"localhost" validate:"required"`
//	    Timeout int    `default:"30"`
//	}
//
//	var cfg Config
//	// Sets defaults only
//	if err := fuda.SetDefaults(&cfg); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Sets defaults AND validates
//	if err := fuda.SetDefaults(&cfg, fuda.WithValidation(true)); err != nil {
//	    log.Fatal(err)
//	}
func SetDefaults(target any, opts ...Option) error {
	cfg := &config{
		validate: false, // Default: no validation
	}
	for _, opt := range opts {
		opt(cfg)
	}

	builder := New()
	if !cfg.validate {
		builder.WithValidator(nil)
	} else if cfg.validator != nil {
		builder.WithValidator(cfg.validator)
	}

	l, err := builder.Build()
	if err != nil {
		return err
	}

	return l.Load(target)
}

// MustSetDefaults is like SetDefaults but panics on error.
// Useful for package-level variable initialization.
func MustSetDefaults(target any, opts ...Option) {
	if err := SetDefaults(target, opts...); err != nil {
		panic("fuda: " + err.Error())
	}
}

// MustLoadFile is like LoadFile but panics on error.
// Useful for package-level variable initialization.
func MustLoadFile(path string, target any) {
	if err := LoadFile(path, target); err != nil {
		panic("fuda: " + err.Error())
	}
}

// MustLoadBytes is like LoadBytes but panics on error.
// Useful for package-level variable initialization.
func MustLoadBytes(data []byte, target any) {
	if err := LoadBytes(data, target); err != nil {
		panic("fuda: " + err.Error())
	}
}

// MustLoadReader is like LoadReader but panics on error.
// Useful for package-level variable initialization.
func MustLoadReader(r io.Reader, target any) {
	if err := LoadReader(r, target); err != nil {
		panic("fuda: " + err.Error())
	}
}

// Validate runs validation on target using the `validate` tag.
// No loading, default processing, or env resolution occurs.
// Only validation is performed.
func Validate(target any, opts ...Option) error {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	v := cfg.validator
	if v == nil {
		v = validator.New()
	}

	return v.Struct(target)
}

// LoadEnv applies environment variables to target via `env` tags.
// No file source is read and no defaults are applied.
// Only env tag processing occurs.
func LoadEnv(target any) error {
	return LoadEnvWithPrefix("", target)
}

// LoadEnvWithPrefix is like LoadEnv but prepends prefix to env var names.
// For example, with prefix "APP_", an `env:"HOST"` tag reads APP_HOST.
func LoadEnvWithPrefix(prefix string, target any) error {
	l, err := New().WithEnvPrefix(prefix).WithValidator(nil).Build()
	if err != nil {
		return err
	}

	return l.Load(target)
}
