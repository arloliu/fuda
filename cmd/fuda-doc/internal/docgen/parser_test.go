package docgen_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen"
)

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}

	return filepath.Join(filepath.Dir(file), "testdata")
}

// ---------- ParsePackage / FindStruct ----------------------------------

func TestFindStruct_Found(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	for _, name := range []string{"Config", "Flat", "WithPointer", "DeepNest"} {
		if ts := p.FindStruct(pkg, name); ts == nil {
			t.Errorf("FindStruct(%q) = nil, want non-nil", name)
		}
	}
}

func TestFindStruct_NotFound(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	if ts := p.FindStruct(pkg, "NonExistent"); ts != nil {
		t.Errorf("FindStruct(NonExistent) = %v, want nil", ts.Name.Name)
	}
}

func TestFindStruct_NonStructTypeAlias(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	// Duration is `type Duration int64` — not a struct, should not be found.
	if ts := p.FindStruct(pkg, "Duration"); ts != nil {
		t.Error("FindStruct(Duration) should return nil for non-struct type alias")
	}
}

// ---------- FindAllStructs --------------------------------------------

func TestFindAllStructs_OnlyExportedStructs(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	all := p.FindAllStructs(pkg)
	if len(all) == 0 {
		t.Fatal("FindAllStructs returned 0 structs")
	}

	nameSet := make(map[string]bool)
	for _, ts := range all {
		nameSet[ts.Name.Name] = true
	}

	// Must include exported structs.
	for _, want := range []string{"Config", "Flat", "WithPointer", "DeepNest", "NoTags", "NoComments"} {
		if !nameSet[want] {
			t.Errorf("expected %q in FindAllStructs result", want)
		}
	}

	// Must NOT include unexported structs or non-struct type aliases.
	for _, bad := range []string{"unexportedConfig", "Duration"} {
		if nameSet[bad] {
			t.Errorf("FindAllStructs should NOT include %q", bad)
		}
	}
}

// ---------- Doc propagation -------------------------------------------

func TestProcessStruct_DocComment(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "Config")
	if ts == nil {
		t.Fatal("Config not found")
	}

	if ts.Doc == nil {
		t.Fatal("Config doc comment is nil — propagateDoc not working")
	}

	doc := ts.Doc.Text()
	if doc == "" {
		t.Error("Config doc comment text is empty")
	}
}

// ---------- Flat struct (scalar fields, tags) --------------------------

func TestProcessStruct_FlatFields(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "Flat")
	if ts == nil {
		t.Fatal("Flat not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(Flat): %v", err)
	}

	assertFieldCount(t, "Flat", fields, 3)

	// Name
	assertField(t, fields[0], "Name", "string", map[string]string{
		"yaml":    "name",
		"default": "flat",
		"env":     "FLAT_NAME",
	})
	if fields[0].Description == "" {
		t.Error("Flat.Name should have a doc comment")
	}

	// Count
	assertField(t, fields[1], "Count", "int", map[string]string{
		"yaml":    "count",
		"default": "10",
	})

	// Enabled
	assertField(t, fields[2], "Enabled", "bool", map[string]string{
		"yaml":    "enabled",
		"default": "true",
		"env":     "FLAT_ENABLED",
	})
}

// ---------- Same-package nested struct via pointer ---------------------

func TestProcessStruct_PointerNested(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "WithPointer")
	if ts == nil {
		t.Fatal("WithPointer not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(WithPointer): %v", err)
	}

	assertFieldCount(t, "WithPointer", fields, 2)

	inner := fields[1]
	if inner.Name != "Inner" {
		t.Fatalf("expected field Inner, got %s", inner.Name)
	}

	if inner.Type != "*InnerConfig" {
		t.Errorf("Inner type = %q, want *InnerConfig", inner.Type)
	}

	if inner.NestedType != "InnerConfig" {
		t.Errorf("Inner.NestedType = %q, want InnerConfig", inner.NestedType)
	}

	assertFieldCount(t, "InnerConfig", inner.Nested, 2)
	assertField(t, inner.Nested[0], "Value", "string", map[string]string{
		"yaml":    "value",
		"default": "inner-default",
		"env":     "INNER_VALUE",
	})
}

// ---------- Deep nesting (3 levels) -----------------------------------

func TestProcessStruct_DeepNesting(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "DeepNest")
	if ts == nil {
		t.Fatal("DeepNest not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(DeepNest): %v", err)
	}

	// DeepNest -> Level1 -> Level2 -> Level3
	assertFieldCount(t, "DeepNest", fields, 1)

	l1 := fields[0]
	if l1.NestedType != "Level1Config" {
		t.Fatalf("Level1 NestedType = %q", l1.NestedType)
	}

	assertFieldCount(t, "Level1Config", l1.Nested, 2) // Name + Level2

	l2 := l1.Nested[1]
	if l2.NestedType != "Level2Config" {
		t.Fatalf("Level2 NestedType = %q", l2.NestedType)
	}

	assertFieldCount(t, "Level2Config", l2.Nested, 2) // Name + Level3

	l3 := l2.Nested[1]
	if l3.NestedType != "Level3Config" {
		t.Fatalf("Level3 NestedType = %q", l3.NestedType)
	}

	assertFieldCount(t, "Level3Config", l3.Nested, 1)
	assertField(t, l3.Nested[0], "Value", "string", map[string]string{
		"yaml":    "value",
		"default": "deep",
	})
}

// ---------- Embedded struct -------------------------------------------

func TestProcessStruct_Embedded(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "WithEmbedded")
	if ts == nil {
		t.Fatal("WithEmbedded not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(WithEmbedded): %v", err)
	}

	assertFieldCount(t, "WithEmbedded", fields, 2) // Visible + EmbeddedMeta

	embedded := fields[1]
	if embedded.Name != "EmbeddedMeta" {
		t.Fatalf("expected embedded EmbeddedMeta, got %s", embedded.Name)
	}

	if embedded.NestedType != "EmbeddedMeta" {
		t.Errorf("NestedType = %q, want EmbeddedMeta", embedded.NestedType)
	}

	assertFieldCount(t, "EmbeddedMeta", embedded.Nested, 2) // Version + Author
}

// ---------- Slice and map fields --------------------------------------

func TestProcessStruct_SliceAndMapFields(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "WithSliceAndMap")
	if ts == nil {
		t.Fatal("WithSliceAndMap not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(WithSliceAndMap): %v", err)
	}

	assertFieldCount(t, "WithSliceAndMap", fields, 3)
	assertField(t, fields[0], "Items", "[]string", map[string]string{
		"yaml":    "items",
		"default": "a,b,c",
	})
	assertField(t, fields[1], "Labels", "map[string]string", map[string]string{
		"yaml":    "labels",
		"default": "env:dev,tier:web",
	})
	assertField(t, fields[2], "Flags", "map[string]bool", map[string]string{
		"yaml": "flags",
	})

	// Slice and map fields should NOT have nested children.
	for _, f := range fields {
		if len(f.Nested) > 0 {
			t.Errorf("field %s has Nested children, want none for slice/map", f.Name)
		}
	}
}

// ---------- All tag types exercised -----------------------------------

func TestProcessStruct_AllTags(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "WithAllTags")
	if ts == nil {
		t.Fatal("WithAllTags not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(WithAllTags): %v", err)
	}

	assertFieldCount(t, "WithAllTags", fields, 4)

	// FieldA — yaml, default, env, validate, ref, json
	a := fields[0]
	for _, k := range []string{"yaml", "default", "env", "validate", "ref", "json"} {
		if a.Tags[k] == "" {
			t.Errorf("FieldA missing tag %q", k)
		}
	}

	// FieldB — refFrom
	b := fields[1]
	if b.Tags["refFrom"] != "FieldBPath" {
		t.Errorf("FieldB refFrom = %q, want FieldBPath", b.Tags["refFrom"])
	}

	// FieldC — dsn
	c := fields[3]
	if c.Tags["dsn"] == "" {
		t.Error("FieldC missing dsn tag")
	}
}

// ---------- No tags / no comments -------------------------------------

func TestProcessStruct_NoTags(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "NoTags")
	if ts == nil {
		t.Fatal("NoTags not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(NoTags): %v", err)
	}

	assertFieldCount(t, "NoTags", fields, 2)

	for _, f := range fields {
		if len(f.Tags) != 0 {
			t.Errorf("field %s should have no tags, got %v", f.Name, f.Tags)
		}
		if f.Description == "" {
			t.Errorf("field %s should still have a doc comment", f.Name)
		}
	}
}

func TestProcessStruct_NoComments(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "NoComments")
	if ts == nil {
		t.Fatal("NoComments not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(NoComments): %v", err)
	}

	assertFieldCount(t, "NoComments", fields, 2)

	for _, f := range fields {
		if f.Description != "" {
			t.Errorf("field %s should have no description, got %q", f.Name, f.Description)
		}
		// But tags should still be present.
		if f.Tags["yaml"] == "" {
			t.Errorf("field %s should have yaml tag", f.Name)
		}
	}
}

// ---------- Cross-package struct resolution ----------------------------

func TestProcessStruct_CrossPackageDirect(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "Config")
	if ts == nil {
		t.Fatal("Config not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(Config): %v", err)
	}

	// Find Messaging field (type msg.KafkaConfig — direct cross-package with alias)
	messaging := findField(t, fields, "Messaging")
	if messaging.Type != "msg.KafkaConfig" {
		t.Errorf("Messaging.Type = %q, want msg.KafkaConfig", messaging.Type)
	}

	if messaging.NestedType != "KafkaConfig" {
		t.Errorf("Messaging.NestedType = %q, want KafkaConfig", messaging.NestedType)
	}

	if len(messaging.Nested) == 0 {
		t.Fatal("Messaging.Nested is empty — cross-package resolution failed")
	}

	// KafkaConfig should have: Brokers, ClientID, Consumer, Producer, SASL
	assertMinFieldCount(t, "KafkaConfig", messaging.Nested, 5)

	// Verify nested sub-struct resolution within the cross-package.
	consumer := findField(t, messaging.Nested, "Consumer")
	if consumer.NestedType != "ConsumerConfig" {
		t.Errorf("Consumer.NestedType = %q, want ConsumerConfig", consumer.NestedType)
	}

	if len(consumer.Nested) == 0 {
		t.Error("Consumer.Nested is empty — nested resolution within cross-package failed")
	}
}

func TestProcessStruct_CrossPackagePointer(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "Config")
	if ts == nil {
		t.Fatal("Config not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(Config): %v", err)
	}

	// Find Auth field (type *oa.OAuthConfig — pointer to cross-package struct with alias)
	auth := findField(t, fields, "Auth")
	if auth.Type != "*oa.OAuthConfig" {
		t.Errorf("Auth.Type = %q, want *oa.OAuthConfig", auth.Type)
	}

	if auth.NestedType != "OAuthConfig" {
		t.Errorf("Auth.NestedType = %q, want OAuthConfig", auth.NestedType)
	}

	if len(auth.Nested) == 0 {
		t.Fatal("Auth.Nested is empty — pointer cross-package resolution failed")
	}

	// OAuthConfig should have: Issuer, ClientID, ClientSecret, Scopes, TokenExpiry, Providers
	assertMinFieldCount(t, "OAuthConfig", auth.Nested, 5)
}

func TestProcessStruct_CrossPackageViaWrapper(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "StorageConfig")
	if ts == nil {
		t.Fatal("StorageConfig not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(StorageConfig): %v", err)
	}

	assertFieldCount(t, "StorageConfig", fields, 2)

	// Cassandra — value type, cross-package
	cass := fields[0]
	if cass.NestedType != "CassandraConfig" {
		t.Errorf("Cassandra.NestedType = %q, want CassandraConfig", cass.NestedType)
	}

	if len(cass.Nested) == 0 {
		t.Fatal("Cassandra.Nested is empty — cross-package value type resolution failed")
	}

	// CassandraConfig has nested Auth (*CassandraAuth) and Retry (RetryPolicy)
	cassAuth := findField(t, cass.Nested, "Auth")
	if cassAuth.NestedType != "CassandraAuth" {
		t.Errorf("Cassandra.Auth.NestedType = %q, want CassandraAuth", cassAuth.NestedType)
	}

	if len(cassAuth.Nested) == 0 {
		t.Error("Cassandra.Auth.Nested is empty — nested pointer within cross-package failed")
	}

	retry := findField(t, cass.Nested, "Retry")
	if retry.NestedType != "RetryPolicy" {
		t.Errorf("Cassandra.Retry.NestedType = %q, want RetryPolicy", retry.NestedType)
	}

	if len(retry.Nested) == 0 {
		t.Error("Cassandra.Retry.Nested is empty — nested value within cross-package failed")
	}

	// ObjectStore — pointer type, cross-package
	obj := fields[1]
	if obj.NestedType != "S3Config" {
		t.Errorf("ObjectStore.NestedType = %q, want S3Config", obj.NestedType)
	}

	if len(obj.Nested) == 0 {
		t.Fatal("ObjectStore.Nested is empty — cross-package pointer resolution failed")
	}

	// S3Config has Credentials (S3Credentials)
	creds := findField(t, obj.Nested, "Credentials")
	if creds.NestedType != "S3Credentials" {
		t.Errorf("S3.Credentials.NestedType = %q, want S3Credentials", creds.NestedType)
	}

	if len(creds.Nested) == 0 {
		t.Error("S3.Credentials.Nested is empty")
	}
}

// ---------- ParseAll integration test ---------------------------------

func TestParseAll_SingleStruct(t *testing.T) {
	t.Parallel()

	docs, err := docgen.ParseAll("Flat", testdataDir(t))
	if err != nil {
		t.Fatalf("ParseAll(Flat): %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("ParseAll(Flat) returned %d docs, want 1", len(docs))
	}

	if docs[0].Name != "Flat" {
		t.Errorf("docs[0].Name = %q, want Flat", docs[0].Name)
	}

	if len(docs[0].Fields) != 3 {
		t.Errorf("Flat fields = %d, want 3", len(docs[0].Fields))
	}
}

func TestParseAll_AllStructs(t *testing.T) {
	t.Parallel()

	docs, err := docgen.ParseAll("", testdataDir(t))
	if err != nil {
		t.Fatalf("ParseAll(): %v", err)
	}

	if len(docs) == 0 {
		t.Fatal("ParseAll() returned 0 docs")
	}

	nameSet := make(map[string]bool)
	for _, d := range docs {
		nameSet[d.Name] = true
	}

	for _, want := range []string{"Config", "Flat", "WithPointer", "DeepNest", "WithSliceAndMap"} {
		if !nameSet[want] {
			t.Errorf("expected struct %q in ParseAll results", want)
		}
	}

	// unexported must not appear.
	if nameSet["unexportedConfig"] {
		t.Error("unexportedConfig should not appear in ParseAll")
	}
}

func TestParseAll_NonExistent(t *testing.T) {
	t.Parallel()

	_, err := docgen.ParseAll("DoesNotExist", testdataDir(t))
	if err == nil {
		t.Error("ParseAll(DoesNotExist) should return error")
	}
}

func TestParseAll_InvalidPath(t *testing.T) {
	t.Parallel()

	_, err := docgen.ParseAll("Config", "/nonexistent/path")
	if err == nil {
		t.Error("ParseAll with invalid path should return error")
	}
}

// ---------- Same-package nesting within Config -----------------------

func TestProcessStruct_SamePackageNesting(t *testing.T) {
	t.Parallel()

	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	ts := p.FindStruct(pkg, "Config")
	if ts == nil {
		t.Fatal("Config not found")
	}

	fields, err := p.ProcessStruct(ts, pkg)
	if err != nil {
		t.Fatalf("ProcessStruct(Config): %v", err)
	}

	// Server -> TLS (pointer, same-package, 2 levels)
	server := findField(t, fields, "Server")
	if server.NestedType != "ServerConfig" {
		t.Fatalf("Server.NestedType = %q, want ServerConfig", server.NestedType)
	}

	tls := findField(t, server.Nested, "TLS")
	if tls.NestedType != "TLSConfig" {
		t.Errorf("TLS.NestedType = %q, want TLSConfig", tls.NestedType)
	}

	if len(tls.Nested) != 3 {
		t.Errorf("TLSConfig fields = %d, want 3 (Enabled, CertFile, KeyFile)", len(tls.Nested))
	}

	// Database -> Primary (SQLConfig, value), Analytics (*SQLConfig, pointer)
	db := findField(t, fields, "Database")
	primary := findField(t, db.Nested, "Primary")
	if primary.NestedType != "SQLConfig" {
		t.Errorf("Primary.NestedType = %q, want SQLConfig", primary.NestedType)
	}

	analytics := findField(t, db.Nested, "Analytics")
	if analytics.NestedType != "SQLConfig" {
		t.Errorf("Analytics.NestedType = %q, want SQLConfig", analytics.NestedType)
	}

	if analytics.Type != "*SQLConfig" {
		t.Errorf("Analytics.Type = %q, want *SQLConfig", analytics.Type)
	}

	// Verify Primary and Analytics have same field structure.
	if len(primary.Nested) != len(analytics.Nested) {
		t.Errorf("Primary has %d fields, Analytics has %d — should match",
			len(primary.Nested), len(analytics.Nested))
	}
}

// ---------- Non-struct type not processable ---------------------------

func TestProcessStruct_NonStructError(t *testing.T) {
	t.Parallel()

	// Duration is `type Duration int64` — FindStruct and ParseAll should
	// both refuse to process it.
	docs, err := docgen.ParseAll("Duration", testdataDir(t))
	if err == nil {
		t.Error("ParseAll(Duration) should return error for non-struct type")
	}

	_ = docs

	// Also verify FindStruct directly returns nil.
	p := docgen.NewParser()
	pkg, err := p.ParsePackage(testdataDir(t))
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	if ts := p.FindStruct(pkg, "Duration"); ts != nil {
		t.Error("FindStruct(Duration) should return nil")
	}
}

// ---------- Helpers ---------------------------------------------------

func assertFieldCount(t *testing.T, structName string, fields []docgen.FieldInfo, want int) {
	t.Helper()

	if len(fields) != want {
		names := make([]string, len(fields))
		for i, f := range fields {
			names[i] = f.Name
		}
		t.Fatalf("%s: field count = %d, want %d; fields: %v", structName, len(fields), want, names)
	}
}

func assertMinFieldCount(t *testing.T, structName string, fields []docgen.FieldInfo, minWant int) {
	t.Helper()

	if len(fields) < minWant {
		t.Fatalf("%s: field count = %d, want at least %d", structName, len(fields), minWant)
	}
}

func assertField(t *testing.T, f docgen.FieldInfo, wantName, wantType string, wantTags map[string]string) {
	t.Helper()

	if f.Name != wantName {
		t.Errorf("field name = %q, want %q", f.Name, wantName)
	}

	if f.Type != wantType {
		t.Errorf("field %s type = %q, want %q", wantName, f.Type, wantType)
	}

	for k, v := range wantTags {
		got := f.Tags[k]
		if got != v {
			t.Errorf("field %s tag[%s] = %q, want %q", wantName, k, got, v)
		}
	}
}

func findField(t *testing.T, fields []docgen.FieldInfo, name string) docgen.FieldInfo {
	t.Helper()

	for _, f := range fields {
		if f.Name == name {
			return f
		}
	}

	available := make([]string, len(fields))
	for i, f := range fields {
		available[i] = f.Name
	}

	t.Fatalf("field %q not found; available: %v", name, available)

	return docgen.FieldInfo{}
}
