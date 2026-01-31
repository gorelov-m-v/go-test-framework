package contract

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	specCache   = make(map[string]*openapi3.T)
	specCacheMu sync.RWMutex
)

func Load(specPath string) (*openapi3.T, error) {
	absPath, err := resolveSpecPath(specPath)
	if err != nil {
		return nil, err
	}

	specCacheMu.RLock()
	if spec, ok := specCache[absPath]; ok {
		specCacheMu.RUnlock()
		return spec, nil
	}
	specCacheMu.RUnlock()

	specCacheMu.Lock()
	defer specCacheMu.Unlock()

	if spec, ok := specCache[absPath]; ok {
		return spec, nil
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec from '%s': %w", absPath, err)
	}

	specCache[absPath] = spec
	return spec, nil
}

// MustLoad loads OpenAPI spec or panics on error.
// Use ONLY in init() or package-level var initialization.
// For runtime loading, use Load() which returns an error.
//
// Example:
//
//	var spec = contract.MustLoad("openapi/api.yaml") // OK: package-level
//
//	func init() {
//	    spec = contract.MustLoad("openapi/api.yaml") // OK: init()
//	}
//
//	func handler() {
//	    spec := contract.MustLoad("openapi/api.yaml") // BAD: use Load() instead
//	}
func MustLoad(specPath string) *openapi3.T {
	spec, err := Load(specPath)
	if err != nil {
		panic(fmt.Errorf("contract.MustLoad: %w (use Load() for runtime loading)", err))
	}
	return spec
}

func resolveSpecPath(specPath string) (string, error) {
	if filepath.IsAbs(specPath) {
		if _, err := os.Stat(specPath); err != nil {
			return "", fmt.Errorf("spec file not found: %s", specPath)
		}
		return specPath, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	searchPaths := []string{
		filepath.Join(cwd, specPath),
	}

	dir := cwd
	for i := 0; i < 10; i++ {
		openapiDir := filepath.Join(dir, "openapi", filepath.Base(specPath))
		searchPaths = append(searchPaths, openapiDir)

		specDir := filepath.Join(dir, specPath)
		if specDir != searchPaths[0] {
			searchPaths = append(searchPaths, specDir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return filepath.Abs(p)
		}
	}

	return "", fmt.Errorf("spec file not found: %s (searched in %v)", specPath, searchPaths)
}

func ClearCache() {
	specCacheMu.Lock()
	defer specCacheMu.Unlock()
	specCache = make(map[string]*openapi3.T)
}
