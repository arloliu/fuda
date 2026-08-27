# Render a configuration template

Use `WithTemplate` when one configuration file needs values that your program already knows at startup.
Fuda renders the Go template before it parses YAML.

```go
type TemplateData struct {
    Environment string
    Region      string
}

loader, err := fuda.New().
    FromFile("config.yaml").
    WithTemplate(TemplateData{
        Environment: "production",
        Region:      "us-east-1",
    }).
    Build()
```

Write template expressions in the YAML file:

```yaml
environment: "{{ .Environment }}"
endpoint: "https://api.{{ .Region }}.example.com"
```

Quote template output when YAML should treat it as a string.
If your file needs literal `{{` or `}}`, pass custom delimiters with `fuda.WithDelimiters`.

## Order with overrides

Fuda renders the template first.
It applies `WithOverrides` to that rendered source next, then decodes YAML.
An `env` tag can still override the decoded field after that step.

## Next step

[Apply a programmatic override](overrides.md) to a rendered configuration.
