package docgen

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// FieldInfo is an alias for docutil.FieldInfo provided for backwards
// compatibility so existing code in this package can use the unqualified name.
type FieldInfo = docutil.FieldInfo

// Parser handles parsing of Go source files.
//
//nolint:staticcheck // ast.Package used for simplicity, migration to types checker deferred
type Parser struct {
	fset    *token.FileSet
	pkgDirs map[string]*ast.Package // cache: directory → parsed package
	srcDir  string                  // source directory for resolving imports
}

// NewParser creates a new Parser.
//
//nolint:staticcheck // ast.Package used for simplicity
func NewParser() *Parser {
	return &Parser{
		fset:    token.NewFileSet(),
		pkgDirs: make(map[string]*ast.Package),
	}
}

// ParsePackage parses the directory containing the Go files.
//
//nolint:staticcheck // ast.Package used for simplicity, migration to types checker deferred
func (p *Parser) ParsePackage(path string) (*ast.Package, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var dir string
	var preferredPkg string
	if !fileInfo.IsDir() {
		dir = filepath.Dir(path)

		// If a specific file was passed, prefer its package name.
		if f, err := parser.ParseFile(p.fset, path, nil, parser.PackageClauseOnly); err == nil {
			preferredPkg = f.Name.Name
		}
	} else {
		dir = path
	}

	// Record source directory for import resolution (must be absolute).
	if p.srcDir == "" {
		if abs, err := filepath.Abs(dir); err == nil {
			p.srcDir = abs
		} else {
			p.srcDir = dir
		}
	}

	return p.parseDirWithPreferred(dir, preferredPkg)
}

// parseDir parses a directory and caches the result.
//
//nolint:staticcheck // ast.Package used for simplicity
func (p *Parser) parseDir(dir string) (*ast.Package, error) {
	return p.parseDirWithPreferred(dir, "")
}

// parseDirWithPreferred parses a directory and caches the result,
// preferring the specified package name when provided.
//
//nolint:staticcheck // ast.Package used for simplicity
func (p *Parser) parseDirWithPreferred(dir string, preferredPkg string) (*ast.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	if cached, ok := p.pkgDirs[absDir]; ok {
		return cached, nil
	}

	pkgs, err := parser.ParseDir(p.fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	if preferredPkg != "" {
		if pkg, ok := pkgs[preferredPkg]; ok {
			p.pkgDirs[absDir] = pkg

			return pkg, nil
		}
	}

	// We simply pick the first package we find that isn't a test package if possible,
	// or matches the file's package if a specific file was given.
	// For simplicity, we just look for a non-test package first.
	keys := make([]string, 0, len(pkgs))
	for name := range pkgs {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		if !strings.HasSuffix(name, "_test") {
			p.pkgDirs[absDir] = pkgs[name]

			return pkgs[name], nil
		}
	}

	// If only test packages exist or just one package
	if len(keys) > 0 {
		pkg := pkgs[keys[0]]
		p.pkgDirs[absDir] = pkg

		return pkg, nil
	}

	return nil, fmt.Errorf("no packages found in %s", dir)
}

// FindStruct finds a struct definition by name in the package.
// It propagates doc comments from the enclosing GenDecl to the TypeSpec
// when the TypeSpec's own Doc field is nil (standard Go AST behaviour for
// top-level type declarations).
//
//nolint:staticcheck // ast.Package used for simplicity
func (p *Parser) FindStruct(pkg *ast.Package, structName string) *ast.TypeSpec {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}

			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != structName {
					continue
				}

				// Only return if the underlying type is a struct.
				if _, isStruct := ts.Type.(*ast.StructType); !isStruct {
					return nil
				}

				propagateDoc(gd, ts)

				return ts
			}
		}
	}

	return nil
}

// FindAllStructs returns all exported struct type declarations in the package.
// Doc comments are propagated from GenDecl to TypeSpec where needed.
//
//nolint:staticcheck // ast.Package used for simplicity
func (p *Parser) FindAllStructs(pkg *ast.Package) []*ast.TypeSpec {
	var results []*ast.TypeSpec

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}

			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if _, isStruct := ts.Type.(*ast.StructType); !isStruct {
					continue
				}

				if !docutil.IsExported(ts.Name.Name) {
					continue
				}

				propagateDoc(gd, ts)
				results = append(results, ts)
			}
		}
	}

	return results
}

// propagateDoc copies the GenDecl doc comment to the TypeSpec when the
// TypeSpec's own Doc is nil. This handles the common case where a standalone
// `type Foo struct { ... }` has its comment attached to the GenDecl.
func propagateDoc(gd *ast.GenDecl, ts *ast.TypeSpec) {
	if ts.Doc == nil && gd.Doc != nil && len(gd.Specs) == 1 {
		ts.Doc = gd.Doc
	}
}

// ProcessStruct extracts field information from a struct type.
//
//nolint:staticcheck // ast.Package used for simplicity
func (p *Parser) ProcessStruct(ts *ast.TypeSpec, pkg *ast.Package) ([]FieldInfo, error) {
	stack := make(map[string]bool)

	return p.processStructVisited(ts, pkg, stack)
}

//
//nolint:staticcheck // ast.Package used for simplicity, migration to types checker deferred
func (p *Parser) processStructVisited(ts *ast.TypeSpec, pkg *ast.Package, stack map[string]bool) ([]FieldInfo, error) {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("%s is not a struct", ts.Name.Name)
	}

	if key := p.structKey(ts, pkg); key != "" {
		if stack[key] {
			return nil, nil
		}
		stack[key] = true
		defer delete(stack, key)
	}

	var fields []FieldInfo
	for _, field := range st.Fields.List {
		// Handle embedded fields or named fields
		var names []string
		if len(field.Names) == 0 {
			// Embedded field
			names = []string{getTypeName(field.Type)}
		} else {
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
		}

		for _, name := range names {
			info := FieldInfo{
				Name:        name,
				Type:        getTypeName(field.Type),
				Description: getDoc(field.Doc, field.Comment),
				Tags:        parseTags(field.Tag),
			}

			// Check for nested struct (same package or cross-package).
			nestedType, nestedPkg := p.resolveNestedType(field.Type, pkg)
			if nestedType != nil {
				info.NestedType = nestedType.Name.Name

				if key := p.structKey(nestedType, nestedPkg); key == "" || !stack[key] {
					nestedFields, err := p.processStructVisited(nestedType, nestedPkg, stack)
					if err != nil {
						return nil, err
					}
					info.Nested = nestedFields
				}
			}

			fields = append(fields, info)
		}
	}

	return fields, nil
}

//
//nolint:staticcheck // ast.Package used for simplicity, migration to types checker deferred
func (p *Parser) structKey(ts *ast.TypeSpec, pkg *ast.Package) string {
	if ts == nil || ts.Name == nil {
		return ""
	}

	return p.packageKey(pkg) + "::" + ts.Name.Name
}

//
//nolint:staticcheck // ast.Package used for simplicity, migration to types checker deferred
func (p *Parser) packageKey(pkg *ast.Package) string {
	if pkg == nil {
		return ""
	}

	for name := range pkg.Files {
		return filepath.Dir(name)
	}

	return pkg.Name
}

// resolveNestedType resolves a field's type to a struct TypeSpec and the
// package it belongs to. It handles same-package types, pointer types, and
// cross-package selector expressions (e.g., cassandra.Config).
//
//nolint:staticcheck // ast.Package used for simplicity
func (p *Parser) resolveNestedType(expr ast.Expr, pkg *ast.Package) (*ast.TypeSpec, *ast.Package) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Same-package type.
		ts := findStructInPkg(t.Name, pkg)
		if ts != nil {
			return ts, pkg
		}

		return nil, nil

	case *ast.StarExpr:
		// Pointer — unwrap and recurse.
		return p.resolveNestedType(t.X, pkg)

	case *ast.SelectorExpr:
		// Cross-package: pkgAlias.TypeName
		pkgIdent, ok := t.X.(*ast.Ident)
		if !ok {
			return nil, nil
		}

		importPath := findImportPath(pkg, pkgIdent.Name)
		if importPath == "" {
			return nil, nil
		}

		importedPkg := p.resolveImport(importPath)
		if importedPkg == nil {
			return nil, nil
		}

		ts := findStructInPkg(t.Sel.Name, importedPkg)
		if ts != nil {
			return ts, importedPkg
		}

		return nil, nil

	default:
		return nil, nil
	}
}

// findStructInPkg searches for an exported struct type declaration by name
// within a single package.
//
//nolint:staticcheck // ast.Package used for simplicity
func findStructInPkg(name string, pkg *ast.Package) *ast.TypeSpec {
	if isStandardType(name) {
		return nil
	}

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}

			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != name {
					continue
				}

				if _, isStruct := ts.Type.(*ast.StructType); !isStruct {
					return nil // found type but it's not a struct
				}

				propagateDoc(gd, ts)

				return ts
			}
		}
	}

	return nil
}

// findImportPath looks up the import path for a given package alias across
// all files in the package.
//
//nolint:staticcheck // ast.Package used for simplicity
func findImportPath(pkg *ast.Package, alias string) string {
	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)

			// Explicit alias.
			if imp.Name != nil && imp.Name.Name == alias {
				return path
			}

			// Default alias = last element of import path.
			parts := strings.Split(path, "/")
			if parts[len(parts)-1] == alias {
				return path
			}
		}
	}

	return ""
}

// resolveImport resolves an import path to a parsed *ast.Package.
// It first tries go/build, then falls back to module-aware resolution
// by locating go.mod and computing the package directory from the module path.
//
//nolint:staticcheck // ast.Package used for simplicity
func (p *Parser) resolveImport(importPath string) *ast.Package {
	// Try go/build first (works for GOPATH and some module cases).
	bpkg, err := build.Import(importPath, p.srcDir, build.FindOnly)
	if err == nil {
		if pkg, parseErr := p.parseDir(bpkg.Dir); parseErr == nil {
			return pkg
		}
	}

	// Fall back to module-aware resolution.
	dir := p.resolveModuleImport(importPath)
	if dir == "" {
		return nil
	}

	pkg, err := p.parseDir(dir)
	if err != nil {
		return nil
	}

	return pkg
}

// resolveModuleImport locates a package directory by finding the enclosing
// go.mod, extracting its module path, and computing the sub-directory.
func (p *Parser) resolveModuleImport(importPath string) string {
	modRoot, modPath := findGoMod(p.srcDir)
	if modRoot == "" {
		return ""
	}

	// The import must be under the same module.
	if !strings.HasPrefix(importPath, modPath) {
		return ""
	}

	// Trim module path to get relative directory.
	rel := strings.TrimPrefix(importPath, modPath)
	rel = strings.TrimPrefix(rel, "/")

	return filepath.Join(modRoot, rel)
}

// findGoMod walks up from dir to find go.mod and returns (moduleRoot, modulePath).
func findGoMod(dir string) (moduleRoot string, modulePath string) {
	cur := dir
	for {
		modFile := filepath.Join(cur, "go.mod")
		data, err := os.ReadFile(modFile)
		if err == nil {
			modPath := extractModulePath(string(data))
			if modPath != "" {
				return cur, modPath
			}
		}

		parent := filepath.Dir(cur)
		if parent == cur {
			return "", ""
		}
		cur = parent
	}
}

// extractModulePath parses "module <path>" from go.mod content.
func extractModulePath(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}

	return ""
}

func isStandardType(t string) bool {
	switch t {
	case "string", "bool", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune", "float32", "float64", "complex64", "complex128", "error", "time.Duration", "time.Time":
		return true
	}
	return false
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.SelectorExpr:
		return getTypeName(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeName(t.Key) + "]" + getTypeName(t.Value)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func getDoc(doc *ast.CommentGroup, comment *ast.CommentGroup) string {
	var sb strings.Builder
	if doc != nil {
		sb.WriteString(doc.Text())
	}
	if comment != nil {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(comment.Text())
	}

	return strings.TrimSpace(sb.String())
}

var supportedTags = []string{
	"default", "env", "validate", "yaml", "json", "ref", "refFrom", "dsn", "required",
}

func parseTags(tag *ast.BasicLit) map[string]string {
	if tag == nil {
		return nil
	}
	// reflect.StructTag expects string without backticks
	value := strings.Trim(tag.Value, "`")
	tags := make(map[string]string)

	st := reflect.StructTag(value)
	for _, key := range supportedTags {
		if v, ok := st.Lookup(key); ok {
			tags[key] = v
		}
	}

	return tags
}
