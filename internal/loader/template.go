package loader

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplateConfig holds template parsing configuration.
type TemplateConfig struct {
	LeftDelim  string
	RightDelim string
	MissingKey string // "invalid", "zero", "error"
	FuncMap    template.FuncMap
}

// ProcessTemplate applies Go template parsing to the source content.
func ProcessTemplate(source []byte, data any, cfg *TemplateConfig) ([]byte, error) {
	tmpl := template.New("config")

	if cfg != nil {
		if cfg.LeftDelim != "" && cfg.RightDelim != "" {
			tmpl = tmpl.Delims(cfg.LeftDelim, cfg.RightDelim)
		}
		if cfg.MissingKey != "" {
			tmpl = tmpl.Option("missingkey=" + cfg.MissingKey)
		}
		if cfg.FuncMap != nil {
			tmpl = tmpl.Funcs(cfg.FuncMap)
		}
	}

	parsed, err := tmpl.Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := parsed.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("template execution error: %w", err)
	}

	return buf.Bytes(), nil
}
