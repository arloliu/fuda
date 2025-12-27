package loader

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// DotenvConfig holds dotenv file loading configuration.
type DotenvConfig struct {
	Files       []string // Explicit file paths to load
	SearchPaths []string // Directories to search for env file
	SearchName  string   // Filename to search for (e.g., ".env")
	Override    bool     // If true, use godotenv.Overload instead of Load
}

// loadDotenvFiles loads dotenv files based on configuration.
// This is called at the start of Load() before any env tag processing.
func (e *Engine) loadDotenvFiles() error {
	if e.DotenvConfig == nil {
		return nil
	}

	files := e.resolveEnvFiles()
	if len(files) == 0 {
		return nil
	}

	if e.DotenvConfig.Override {
		return godotenv.Overload(files...)
	}

	return godotenv.Load(files...)
}

// resolveEnvFiles returns the list of env files to load.
// Priority: explicit files > search paths
func (e *Engine) resolveEnvFiles() []string {
	if len(e.DotenvConfig.Files) > 0 {
		return filterExistingFiles(e.DotenvConfig.Files)
	}

	if len(e.DotenvConfig.SearchPaths) > 0 && e.DotenvConfig.SearchName != "" {
		return e.searchForEnvFiles()
	}

	return nil
}

// filterExistingFiles returns only files that exist on disk.
// Missing files are silently ignored to support optional .env.local patterns.
func filterExistingFiles(files []string) []string {
	var existing []string
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			existing = append(existing, f)
		}
	}

	return existing
}

// searchForEnvFiles searches for the configured env file in search paths.
// Returns the first file found, or nil if none found.
func (e *Engine) searchForEnvFiles() []string {
	for _, dir := range e.DotenvConfig.SearchPaths {
		path := filepath.Join(dir, e.DotenvConfig.SearchName)
		if _, err := os.Stat(path); err == nil {
			return []string{path}
		}
	}

	return nil
}
