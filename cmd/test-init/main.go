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
    -with-grpc         Include gRPC DSL support (default: false)
    -with-db           Include Database DSL support (default: false)
    -with-redis        Include Redis DSL support (default: false)
    -with-kafka        Include Kafka DSL support (default: false)
    -all               Include all DSLs (default: false)

Examples:
    # Create minimal project (HTTP only)
    test-init my-api-tests

    # With gRPC and Database support
    test-init -with-grpc -with-db my-api-tests

    # With all DSLs
    test-init -all my-api-tests

    # With custom module name
    test-init -module github.com/company/api-tests my-api-tests
`

func main() {
	moduleName := flag.String("module", "", "Go module name (default: project-name)")
	withExample := flag.Bool("with-example", true, "Generate example HTTP client and test")
	withGRPC := flag.Bool("with-grpc", false, "Include gRPC DSL support")
	withDB := flag.Bool("with-db", false, "Include Database DSL support")
	withRedis := flag.Bool("with-redis", false, "Include Redis DSL support")
	withKafka := flag.Bool("with-kafka", false, "Include Kafka DSL support")
	withAll := flag.Bool("all", false, "Include all DSLs")

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

	// If -all flag is set, enable all DSLs
	if *withAll {
		*withGRPC = true
		*withDB = true
		*withRedis = true
		*withKafka = true
	}

	opts := scaffold.Options{
		WithExample: *withExample,
		WithGRPC:    *withGRPC,
		WithDB:      *withDB,
		WithRedis:   *withRedis,
		WithKafka:   *withKafka,
	}

	fmt.Printf("Creating project: %s\n", projectName)
	fmt.Printf("Module: %s\n", *moduleName)
	fmt.Printf("DSLs: HTTP")
	if opts.WithGRPC {
		fmt.Print(", gRPC")
	}
	if opts.WithDB {
		fmt.Print(", Database")
	}
	if opts.WithRedis {
		fmt.Print(", Redis")
	}
	if opts.WithKafka {
		fmt.Print(", Kafka")
	}
	fmt.Println()
	fmt.Println()

	gen := scaffold.NewGenerator(projectName, *moduleName, opts)

	if err := gen.Generate(); err != nil {
		log.Fatalf("Failed to generate project: %v", err)
	}

	fmt.Println("Project created successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  go test ./tests/... -v")

	if opts.WithGRPC {
		fmt.Println()
		fmt.Println("To generate gRPC clients:")
		fmt.Println("  grpc-gen -pb-import \"your-project/pb\" proto/player.proto")
	}

	fmt.Println()
	fmt.Println("To generate HTTP clients from OpenAPI:")
	fmt.Println("  openapi-gen openapi/spec.json")

	absPath, _ := filepath.Abs(projectName)
	fmt.Printf("\nProject location: %s\n", absPath)
}
