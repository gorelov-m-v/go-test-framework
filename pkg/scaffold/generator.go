package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// Options configures which DSLs to include in the project
type Options struct {
	WithExample bool
	WithGRPC    bool
	WithDB      bool
	WithRedis   bool
	WithKafka   bool
}

// Generator creates new test project structure
type Generator struct {
	projectName string
	moduleName  string
	options     Options
}

// NewGenerator creates a new project generator
func NewGenerator(projectName, moduleName string, opts Options) *Generator {
	return &Generator{
		projectName: projectName,
		moduleName:  moduleName,
		options:     opts,
	}
}

// Generate creates the project structure
func (g *Generator) Generate() error {
	// Base directories (always created)
	dirs := []string{
		"configs",
		"internal/client",
		"tests",
		"openapi",
	}

	// Optional directories based on selected DSLs
	if g.options.WithGRPC {
		dirs = append(dirs, "internal/grpc_client", "proto")
	}
	if g.options.WithDB {
		dirs = append(dirs, "internal/db")
	}
	if g.options.WithRedis {
		dirs = append(dirs, "internal/redis")
	}
	if g.options.WithKafka {
		dirs = append(dirs, "internal/kafka")
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
		"go.mod":                    g.goModTemplate(),
		".gitignore":                gitignoreTemplate(),
		"Makefile":                  makefileTemplate(),
		"configs/config.local.yaml": g.configTemplate(),
		"tests/env.go":              g.envTemplate(),
	}

	// Add example if requested
	if g.options.WithExample {
		files["tests/example_test.go"] = g.exampleTestTemplate()
		files["internal/client/jsonplaceholder/client.go"] = g.exampleClientTemplate()
		files["internal/client/jsonplaceholder/models.go"] = g.exampleModelsTemplate()
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
	gitkeepDirs := []string{"openapi"}

	if g.options.WithGRPC {
		gitkeepDirs = append(gitkeepDirs, "internal/grpc_client", "proto")
	}
	if g.options.WithDB {
		gitkeepDirs = append(gitkeepDirs, "internal/db")
	}
	if g.options.WithRedis {
		gitkeepDirs = append(gitkeepDirs, "internal/redis")
	}
	if g.options.WithKafka {
		gitkeepDirs = append(gitkeepDirs, "internal/kafka")
	}
	if !g.options.WithExample {
		gitkeepDirs = append(gitkeepDirs, "internal/client")
	}

	for _, dir := range gitkeepDirs {
		gitkeepPath := filepath.Join(g.projectName, dir, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to write .gitkeep in %s: %w", dir, err)
		}
	}

	return nil
}
