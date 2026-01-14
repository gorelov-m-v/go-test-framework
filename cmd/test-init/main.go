package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gorelov-m-v/go-test-framework/pkg/scaffold"
)

const usage = `Go Test Framework Project Generator

Usage:
    test-init [options] <project-name>

Arguments:
    project-name    Name of the project directory to create

Options:
    -module string     Go module name (default: project-name)
    -with-example      Generate example HTTP client and test (default: true)

Examples:
    # Create new project
    test-init my-api-tests

    # With custom module name
    test-init -module github.com/company/api-tests my-api-tests

    # Minimal project without examples
    test-init -with-example=false my-api-tests

Generated structure:
    my-api-tests/
    ├── configs/
    │   └── config.local.yaml
    ├── internal/
    │   ├── http_client/
    │   ├── grpc_client/
    │   ├── db/
    │   ├── redis/
    │   └── kafka/
    ├── tests/
    │   ├── env.go
    │   └── example_test.go
    ├── Makefile
    ├── .gitignore
    └── go.mod
`

func main() {
	moduleName := flag.String("module", "", "Go module name (default: project-name)")
	withExample := flag.Bool("with-example", true, "Generate example HTTP client and test")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	projectName := flag.Arg(0)

	if *moduleName == "" {
		*moduleName = projectName
	}

	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		log.Fatalf("Directory '%s' already exists", projectName)
	}

	fmt.Printf("Creating project: %s\n", projectName)
	fmt.Printf("Module: %s\n", *moduleName)
	fmt.Println()

	gen := scaffold.NewGenerator(projectName, *moduleName, *withExample)

	if err := gen.Generate(); err != nil {
		log.Fatalf("Failed to generate project: %v", err)
	}

	fmt.Println("Project created successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  go test ./tests/... -v")
	fmt.Println()
	fmt.Println("To generate clients from OpenAPI/Proto:")
	fmt.Println("  openapi-gen openapi.json")
	fmt.Println("  grpc-gen -pb-import \"your-project/pb\" player.proto")

	absPath, _ := filepath.Abs(projectName)
	fmt.Printf("\nProject location: %s\n", absPath)
}
