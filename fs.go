package fuda

import "github.com/spf13/afero"

// DefaultFs is the default filesystem used by fuda for all file operations.
// It defaults to the OS filesystem but can be overridden for testing.
//
// Example usage for testing:
//
//	func TestMyConfig(t *testing.T) {
//	    memFs := afero.NewMemMapFs()
//	    afero.WriteFile(memFs, "/config.yaml", []byte("host: localhost"), 0644)
//	    fuda.SetDefaultFs(memFs)
//	    defer fuda.ResetDefaultFs()
//	    // ... test code ...
//	}
var DefaultFs afero.Fs = afero.NewOsFs()

// SetDefaultFs sets the global default filesystem.
// This is useful for testing scenarios where all file operations
// should use a memory filesystem or other custom implementation.
//
// WARNING: This modifies global state and is NOT thread-safe.
// Do not use with t.Parallel() tests. For concurrent tests,
// use WithFilesystem() on individual builders instead.
//
// Call during test setup (e.g., TestMain), not during test execution.
func SetDefaultFs(fs afero.Fs) {
	DefaultFs = fs
}

// ResetDefaultFs resets the global filesystem to the OS filesystem.
// Call this in test cleanup to restore default behavior.
func ResetDefaultFs() {
	DefaultFs = afero.NewOsFs()
}
