package tests

import (
	"strings"
	"testing"
	"text/template"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TemplateConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Debug    bool   `yaml:"debug"`
}

type TemplateData struct {
	Host     string
	Port     int
	Database string
	Debug    bool
}

func TestTemplate_Basic(t *testing.T) {
	yamlContent := `
host: "{{ .Host }}"
port: {{ .Port }}
database: "{{ .Database }}"
debug: {{ .Debug }}
`
	data := TemplateData{
		Host:     "localhost",
		Port:     8080,
		Database: "mydb",
		Debug:    true,
	}

	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "mydb", cfg.Database)
	assert.True(t, cfg.Debug)
}

func TestTemplate_CustomDelimiters(t *testing.T) {
	// Use <% %> delimiters to avoid conflict with {{ }} in YAML
	yamlContent := `
host: "<% .Host %>"
port: <% .Port %>
`
	data := TemplateData{
		Host: "example.com",
		Port: 9000,
	}

	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data, fuda.WithDelimiters("<%", "%>")).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "example.com", cfg.Host)
	assert.Equal(t, 9000, cfg.Port)
}

func TestTemplate_CustomFuncs(t *testing.T) {
	yamlContent := `
host: "{{ upper .Host }}"
`
	data := TemplateData{
		Host: "localhost",
	}

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}

	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data, fuda.WithFuncs(funcMap)).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "LOCALHOST", cfg.Host)
}

func TestTemplate_MissingKeyError(t *testing.T) {
	yamlContent := `
host: "{{ .NonExistent }}"
`
	data := TemplateData{
		Host: "localhost",
	}

	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data, fuda.WithMissingKey("error")).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template execution error")
}

func TestTemplate_MissingKeyZero(t *testing.T) {
	// missingkey=zero still outputs <no value> for missing map keys
	// but doesn't error unlike missingkey=error
	// The zero option is mainly useful for struct fields with zero values
	yamlContent := `
host: "{{ .NonExistent }}"
`
	data := map[string]any{
		"Host": "localhost",
	}

	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data, fuda.WithMissingKey("zero")).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	// With zero, missing keys don't error, but output <no value> for maps
	require.NoError(t, err)
	assert.Equal(t, "<no value>", cfg.Host)
}

func TestTemplate_NestedData(t *testing.T) {
	type DBConfig struct {
		User string
		Pass string
	}
	type NestedData struct {
		Host string
		DB   DBConfig
	}

	yamlContent := `
host: "{{ .Host }}"
database: "{{ .DB.User }}:{{ .DB.Pass }}"
`
	data := NestedData{
		Host: "localhost",
		DB: DBConfig{
			User: "admin",
			Pass: "secret",
		},
	}

	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "admin:secret", cfg.Database)
}

func TestTemplate_NoTemplate_BackwardCompatible(t *testing.T) {
	// Without WithTemplate, config should load normally
	yamlContent := `
host: "plainhost"
port: 3000
`
	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "plainhost", cfg.Host)
	assert.Equal(t, 3000, cfg.Port)
}

func TestTemplate_InvalidTemplate(t *testing.T) {
	yamlContent := `
host: "{{ .Host"
`
	data := TemplateData{
		Host: "localhost",
	}

	var cfg TemplateConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template parse error")
}

func TestTemplate_WithDefaults(t *testing.T) {
	type ConfigWithDefaults struct {
		Host    string `yaml:"host" default:"fallback"`
		Port    int    `yaml:"port" default:"8080"`
		Timeout string `yaml:"timeout" default:"30s"`
	}

	yamlContent := `
host: "{{ .Host }}"
`
	data := TemplateData{
		Host: "templated-host",
	}

	var cfg ConfigWithDefaults
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(data).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	// Template value applied
	assert.Equal(t, "templated-host", cfg.Host)
	// Defaults applied for missing fields
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "30s", cfg.Timeout)
}

func TestTemplate_EmptySource(t *testing.T) {
	// No source, just defaults
	type ConfigWithDefaults struct {
		Host string `default:"default-host"`
	}

	data := TemplateData{Host: "ignored"}

	var cfg ConfigWithDefaults
	loader, err := fuda.New().
		WithTemplate(data).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "default-host", cfg.Host)
}
