package tests

import (
	"os"
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- P1: Edge Cases and Error Paths ---

func TestEdgeCase_NilTarget(t *testing.T) {
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-nil pointer")
}

func TestEdgeCase_NonPointerTarget(t *testing.T) {
	type Config struct {
		Host string `default:"localhost"`
	}

	loader, err := fuda.New().Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(cfg) // Not a pointer!
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-nil pointer")
}

func TestEdgeCase_InvalidYAMLSyntax(t *testing.T) {
	type Config struct {
		Host string `yaml:"host"`
	}

	invalidYAML := `
host: "localhost
  port: 8080
`
	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte(invalidYAML)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestEdgeCase_InvalidDefaultValue(t *testing.T) {
	//nolint:revive // intentionally testing invalid default value
	type Config struct {
		Port int `default:"not_a_number"`
	}

	cfg := &Config{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
}

func TestEdgeCase_RefResolutionError(t *testing.T) {
	type Config struct {
		Secret string `ref:"file://nonexistent_file_12345"`
	}

	cfg := &Config{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ref")
}

func TestEdgeCase_UnexportedFieldsSkipped(t *testing.T) {
	type Config struct {
		Public  string `default:"public_value"`
		private string //nolint:unused // intentionally unexported for testing
	}

	cfg := &Config{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "public_value", cfg.Public)
	// private field should be empty (not set)
}

func TestEdgeCase_EnvNotSet(t *testing.T) {
	type Config struct {
		Host string `env:"NONEXISTENT_ENV_VAR_12345" default:"fallback"`
	}

	cfg := &Config{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "fallback", cfg.Host, "Should use default when env not set")
}

func TestEdgeCase_EnvOverridesDefault(t *testing.T) {
	type Config struct {
		Host string `env:"TEST_HOST_EDGE" default:"default_host"`
	}

	require.NoError(t, os.Setenv("TEST_HOST_EDGE", "env_host"))
	defer os.Unsetenv("TEST_HOST_EDGE")

	cfg := &Config{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "env_host", cfg.Host, "Env should override default")
}

func TestEdgeCase_YAMLOverridesDefault(t *testing.T) {
	type Config struct {
		Host string `yaml:"host" default:"default_host"`
	}

	yamlContent := `host: "yaml_host"`
	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "yaml_host", cfg.Host, "YAML should override default")
}

// --- P2: Setter Interface ---

type ConfigWithSetter struct {
	Host        string
	Port        int    `default:"8080"`
	FullAddress string // Computed by SetDefaults
}

func (c *ConfigWithSetter) SetDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}
	// Compute derived value
	c.FullAddress = c.Host + ":" + itoa(c.Port)
}

func itoa(i int) string {
	return string(rune('0'+i/1000%10)) + string(rune('0'+i/100%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i%10))
}

func TestSetter_CalledAfterTagProcessing(t *testing.T) {
	cfg := &ConfigWithSetter{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host, "SetDefaults should set Host")
	assert.Equal(t, 8080, cfg.Port, "Default tag should set Port")
	assert.Equal(t, "localhost:8080", cfg.FullAddress, "SetDefaults should compute FullAddress")
}

func TestSetter_YAMLValueNotOverwritten(t *testing.T) {
	yamlContent := `host: "custom.local"`
	cfg := &ConfigWithSetter{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "custom.local", cfg.Host, "YAML value should be preserved")
	assert.Equal(t, "custom.local:8080", cfg.FullAddress, "SetDefaults should use YAML Host")
}

type NestedWithSetter struct {
	Name   string `default:"nested_default"`
	Suffix string
}

func (n *NestedWithSetter) SetDefaults() {
	n.Suffix = "_processed"
}

type ParentWithNestedSetter struct {
	Prefix string            `default:"prefix"`
	Nested *NestedWithSetter `yaml:"nested"`
}

func TestSetter_NestedStructs(t *testing.T) {
	yamlContent := `
nested:
  name: "custom_name"
`
	cfg := &ParentWithNestedSetter{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "prefix", cfg.Prefix)
	require.NotNil(t, cfg.Nested)
	assert.Equal(t, "custom_name", cfg.Nested.Name)
	assert.Equal(t, "_processed", cfg.Nested.Suffix, "SetDefaults should be called on nested struct")
}

// --- P2: Validation ---

type ValidatedConfig struct {
	Email  string `yaml:"email" validate:"required,email"`
	Port   int    `yaml:"port" validate:"min=1,max=65535"`
	Status string `yaml:"status" validate:"oneof=active inactive"`
}

func TestValidation_RequiredFieldMissing(t *testing.T) {
	yamlContent := `
port: 8080
status: "active"
`
	cfg := &ValidatedConfig{}
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithValidator(validator.New()).
		Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Email")
}

func TestValidation_InvalidEmail(t *testing.T) {
	yamlContent := `
email: "not-an-email"
port: 8080
status: "active"
`
	cfg := &ValidatedConfig{}
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithValidator(validator.New()).
		Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

func TestValidation_PortOutOfRange(t *testing.T) {
	yamlContent := `
email: "test@example.com"
port: 70000
status: "active"
`
	cfg := &ValidatedConfig{}
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithValidator(validator.New()).
		Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Port")
}

func TestValidation_InvalidOneof(t *testing.T) {
	yamlContent := `
email: "test@example.com"
port: 8080
status: "unknown"
`
	cfg := &ValidatedConfig{}
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithValidator(validator.New()).
		Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Status")
}

func TestValidation_AllValid(t *testing.T) {
	yamlContent := `
email: "test@example.com"
port: 8080
status: "active"
`
	cfg := &ValidatedConfig{}
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithValidator(validator.New()).
		Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", cfg.Email)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "active", cfg.Status)
}

// --- Timeout ---

func TestTimeout_Applied(t *testing.T) {
	type Config struct {
		Host string `default:"localhost"`
	}

	cfg := &Config{}
	loader, err := fuda.New().
		WithTimeout(5 * time.Second).
		Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Host)
}

// --- Cycle Detection ---

// SelfReferentialConfig is a struct that references itself, creating a cycle.
type SelfReferentialConfig struct {
	Name string                 `default:"root"`
	Self *SelfReferentialConfig `yaml:"self"`
}

func TestCycleDetection_SelfReferentialStruct(t *testing.T) {
	// Create a self-referential struct
	cfg := &SelfReferentialConfig{}
	cfg.Self = cfg // Create cycle

	// Must disable validator because go-playground/validator also has no cycle detection
	// and will stack overflow on the cyclic validation traversal
	loader, err := fuda.New().WithValidator(nil).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestCycleDetection_IndirectCycle(t *testing.T) {
	// A -> B -> A cycle
	type NodeB struct {
		Name   string `default:"nodeB"`
		Parent any    `yaml:"-"` // Will be set manually
	}
	type NodeA struct {
		Name  string `default:"nodeA"`
		Child *NodeB `yaml:"child"`
	}

	a := &NodeA{}
	b := &NodeB{}
	a.Child = b
	b.Parent = a // Creates indirect cycle, but Parent is `any`, not *NodeA

	// Must disable validator because go-playground/validator also has no cycle detection
	loader, err := fuda.New().WithValidator(nil).Build()
	require.NoError(t, err)

	// This should NOT cause a cycle because Parent is `any` and not processed as struct pointer
	err = loader.Load(a)
	require.NoError(t, err)
	assert.Equal(t, "nodeA", a.Name)
	assert.Equal(t, "nodeB", a.Child.Name)
}
