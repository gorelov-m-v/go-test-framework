package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// Generator creates new test project structure
type Generator struct {
	projectName string
	moduleName  string
	withExample bool
}

// NewGenerator creates a new project generator
func NewGenerator(projectName, moduleName string, withExample bool) *Generator {
	return &Generator{
		projectName: projectName,
		moduleName:  moduleName,
		withExample: withExample,
	}
}

// Generate creates the project structure
func (g *Generator) Generate() error {
	// Create directories
	dirs := []string{
		"configs",
		"internal/http_client",
		"internal/grpc_client",
		"internal/db",
		"internal/redis",
		"internal/kafka",
		"tests",
		"proto",
	}

	for _, dir := range dirs {
		path := filepath.Join(g.projectName, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("  Created: %s/\n", dir)
	}

	// Create files
	files := map[string]string{
		"go.mod":                g.goModTemplate(),
		".gitignore":            gitignoreTemplate(),
		"Makefile":              makefileTemplate(),
		"configs/config.local.yaml": g.configTemplate(),
		"tests/env.go":          g.envTemplate(),
	}

	// Add example if requested
	if g.withExample {
		files["tests/example_test.go"] = g.exampleTestTemplate()
		files["internal/http_client/example/client.go"] = g.exampleClientTemplate()
		files["internal/http_client/example/models.go"] = g.exampleModelsTemplate()
	}

	for path, content := range files {
		fullPath := filepath.Join(g.projectName, path)

		// Ensure directory exists
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
		fmt.Printf("  Created: %s\n", path)
	}

	// Create .gitkeep files for empty directories
	gitkeepDirs := []string{
		"internal/grpc_client",
		"internal/db",
		"internal/redis",
		"internal/kafka",
		"proto",
	}

	if !g.withExample {
		gitkeepDirs = append(gitkeepDirs, "internal/http_client")
	}

	for _, dir := range gitkeepDirs {
		gitkeepPath := filepath.Join(g.projectName, dir, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to write .gitkeep in %s: %w", dir, err)
		}
	}

	return nil
}
