package tests

import (
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PrimitiveStruct covers shared compatible types (Primitives)
type PrimitiveStruct struct {
	String     string        `default:"hello"`
	Int        int           `default:"42"`
	Int8       int8          `default:"8"`
	Int16      int16         `default:"16"`
	Int32      int32         `default:"32"`
	Int64      int64         `default:"64"`
	Uint       uint          `default:"42"`
	Uint8      uint8         `default:"8"`
	Uint16     uint16        `default:"16"`
	Uint32     uint32        `default:"32"`
	Uint64     uint64        `default:"64"`
	Float32    float32       `default:"3.14"`
	Float64    float64       `default:"2.718"`
	BoolTrue   bool          `default:"true"`
	BoolFalse  bool          `default:"false"`
	Duration   time.Duration `default:"1h30m"`
	PtrString  *string       `default:"ptr_val"`
	PtrInt     *int          `default:"100"`
	PtrInt8    *int8         `default:"8"`
	PtrUint    *uint         `default:"42"`
	PtrFloat64 *float64      `default:"2.718"`
	PtrBool    *bool         `default:"true"`

	Nested NestedStruct
}

type NestedStruct struct {
	Field string `default:"nested_default"`
}

// SliceStructFuda uses Comma separated values (Fuda Style)
type SliceStructFuda struct {
	SliceString []string `default:"a,b,c"`
	SliceInt    []int    `default:"1,2,3"`
}

// SliceStructDefaults uses JSON format (Creasty Style)
type SliceStructDefaults struct {
	SliceString []string `default:"[\"a\",\"b\",\"c\"]"`
	SliceInt    []int    `default:"[1,2,3]"`
}

func TestDefaultsParity_Primitives(t *testing.T) {
	// Baseline: creasty/defaults
	expected := PrimitiveStruct{}
	err := defaults.Set(&expected)
	require.NoError(t, err, "creasty/defaults failed")

	// Candidate: fuda
	actual := PrimitiveStruct{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(&actual)
	require.NoError(t, err, "fuda load failed")

	// Compare field by field
	assert.Equal(t, expected.String, actual.String, "String mismatch")
	assert.Equal(t, expected.Int, actual.Int, "Int mismatch")
	assert.Equal(t, expected.Int8, actual.Int8, "Int8 mismatch")
	assert.Equal(t, expected.Int16, actual.Int16, "Int16 mismatch")
	assert.Equal(t, expected.Int32, actual.Int32, "Int32 mismatch")
	assert.Equal(t, expected.Int64, actual.Int64, "Int64 mismatch")
	assert.Equal(t, expected.Uint, actual.Uint, "Uint mismatch")
	assert.Equal(t, expected.Uint8, actual.Uint8, "Uint8 mismatch")
	assert.Equal(t, expected.Uint16, actual.Uint16, "Uint16 mismatch")
	assert.Equal(t, expected.Uint32, actual.Uint32, "Uint32 mismatch")
	assert.Equal(t, expected.Uint64, actual.Uint64, "Uint64 mismatch")
	assert.Equal(t, expected.Float32, actual.Float32, "Float32 mismatch")
	assert.Equal(t, expected.Float64, actual.Float64, "Float64 mismatch")
	assert.Equal(t, expected.BoolTrue, actual.BoolTrue, "BoolTrue mismatch")
	assert.Equal(t, expected.BoolFalse, actual.BoolFalse, "BoolFalse mismatch")
	assert.Equal(t, expected.Duration, actual.Duration, "Duration mismatch")

	// Ptr
	if expected.PtrString != nil && actual.PtrString != nil {
		assert.Equal(t, *expected.PtrString, *actual.PtrString, "PtrString mismatch")
	} else {
		assert.Equal(t, expected.PtrString, actual.PtrString, "PtrString nil mismatch")
	}

	if expected.PtrInt != nil && actual.PtrInt != nil {
		assert.Equal(t, *expected.PtrInt, *actual.PtrInt, "PtrInt mismatch")
	} else {
		assert.Equal(t, expected.PtrInt, actual.PtrInt, "PtrInt nil mismatch")
	}

	if expected.PtrInt8 != nil && actual.PtrInt8 != nil {
		assert.Equal(t, *expected.PtrInt8, *actual.PtrInt8, "PtrInt8 mismatch")
	} else {
		assert.Equal(t, expected.PtrInt8, actual.PtrInt8, "PtrInt8 nil mismatch")
	}

	if expected.PtrUint != nil && actual.PtrUint != nil {
		assert.Equal(t, *expected.PtrUint, *actual.PtrUint, "PtrUint mismatch")
	} else {
		assert.Equal(t, expected.PtrUint, actual.PtrUint, "PtrUint nil mismatch")
	}

	if expected.PtrFloat64 != nil && actual.PtrFloat64 != nil {
		assert.Equal(t, *expected.PtrFloat64, *actual.PtrFloat64, "PtrFloat64 mismatch")
	} else {
		assert.Equal(t, expected.PtrFloat64, actual.PtrFloat64, "PtrFloat64 nil mismatch")
	}

	if expected.PtrBool != nil && actual.PtrBool != nil {
		assert.Equal(t, *expected.PtrBool, *actual.PtrBool, "PtrBool mismatch")
	} else {
		assert.Equal(t, expected.PtrBool, actual.PtrBool, "PtrBool nil mismatch")
	}

	assert.Equal(t, expected.Nested.Field, actual.Nested.Field, "Nested field mismatch")
}

func TestDefaultsParity_Slices(t *testing.T) {
	// Verify that Fuda's comma syntax achieves the same logical result as Defaults' JSON syntax

	// Baseline
	creasty := SliceStructDefaults{}
	err := defaults.Set(&creasty)
	require.NoError(t, err, "creasty/defaults slice failed")

	// Fuda
	fudaVal := SliceStructFuda{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)
	err = loader.Load(&fudaVal)
	require.NoError(t, err, "fuda slice load failed")

	// Compare Logical Result
	assert.Equal(t, creasty.SliceString, fudaVal.SliceString, "SliceString logic mismatch")
	assert.Equal(t, creasty.SliceInt, fudaVal.SliceInt, "SliceInt logic mismatch")
}

func TestSetDefaults_TopLevel(t *testing.T) {
	type SimpleConfig struct {
		Host    string        `default:"localhost"`
		Port    int           `default:"8080"`
		Timeout time.Duration `default:"30s"`
	}

	var cfg SimpleConfig
	err := fuda.SetDefaults(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}
