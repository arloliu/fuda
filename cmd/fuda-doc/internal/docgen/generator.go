package docgen

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// OutputFormat specifies the documentation output format.
type OutputFormat int

const (
	// FormatMarkdown outputs Markdown documentation.
	FormatMarkdown OutputFormat = iota
	// FormatASCII outputs terminal-friendly documentation with ANSI colors.
	FormatASCII
)

// StructDoc holds parsed documentation data for a single struct.
type StructDoc struct {
	Name   string      // struct name
	Doc    string      // struct-level godoc comment
	Fields []FieldInfo // recursive field tree
}

// ParseAll discovers every exported struct in the given path and returns their
// documentation data. When structName is non-empty only that struct is
// returned; when empty all exported structs are included.
func ParseAll(structName, path string) ([]StructDoc, error) {
	parser := NewParser()

	pkg, err := parser.ParsePackage(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package: %w", err)
	}

	if structName != "" {
		ts := parser.FindStruct(pkg, structName)
		if ts == nil {
			return nil, fmt.Errorf("struct %s not found in package", structName)
		}

		fields, err := parser.ProcessStruct(ts, pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to process struct: %w", err)
		}

		doc := ""
		if ts.Doc != nil {
			doc = strings.TrimSpace(ts.Doc.Text())
		}

		return []StructDoc{{Name: structName, Doc: doc, Fields: fields}}, nil
	}

	// Discover all exported structs.
	all := parser.FindAllStructs(pkg)
	if len(all) == 0 {
		return nil, errors.New("no exported structs found in package")
	}

	docs := make([]StructDoc, 0, len(all))

	for _, ts := range all {
		fields, err := parser.ProcessStruct(ts, pkg)
		if err != nil {
			continue // skip structs that can't be processed
		}

		doc := ""
		if ts.Doc != nil {
			doc = strings.TrimSpace(ts.Doc.Text())
		}

		docs = append(docs, StructDoc{Name: ts.Name.Name, Doc: doc, Fields: fields})
	}

	return docs, nil
}

// Generate generates documentation for the specified struct in the given path.
func Generate(structName, path string, w io.Writer, format OutputFormat) error {
	parser := NewParser()

	pkg, err := parser.ParsePackage(path)
	if err != nil {
		return fmt.Errorf("failed to parse package: %w", err)
	}

	ts := parser.FindStruct(pkg, structName)
	if ts == nil {
		return fmt.Errorf("struct %s not found in package", structName)
	}

	fields, err := parser.ProcessStruct(ts, pkg)
	if err != nil {
		return fmt.Errorf("failed to process struct: %w", err)
	}

	doc := ""
	if ts.Doc != nil {
		doc = strings.TrimSpace(ts.Doc.Text())
	}

	switch format {
	case FormatMarkdown:
		printer := NewMarkdownPrinter(w)
		printer.Print(structName, doc, fields)
	case FormatASCII:
		printer := NewASCIIPrinter(w)
		printer.Print(structName, doc, fields)
	default:
		return fmt.Errorf("unsupported output format: %d", format)
	}

	return nil
}
