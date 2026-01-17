package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gorelov-m-v/go-test-framework/internal/codegen/openapi"
)

const usage = `OpenAPI to go-test-framework DSL Generator

Usage:
    openapi-gen [options] <openapi-spec>

Arguments:
    openapi-spec    Path to OpenAPI 3.x specification file (JSON or YAML)

Options:
    -service string    Service name (default: auto-detect from spec)
    -output string     Output directory (default: current directory)
    -client string     Client output path (default: internal/http_client/{service})

Examples:
    # Generate from openapi.json
    openapi-gen openapi.json

    # Specify service name and output directory
    openapi-gen -service auth -output ./generated openapi.yaml

    # Custom path
    openapi-gen -client pkg/http_client/auth openapi.json

Generated files:
    - internal/http_client/{service}/client.go   (Link + DSL methods)
    - internal/http_client/{service}/models.go   (Request/Response types)
`

func main() {
	serviceName := flag.String("service", "", "Service name (default: auto-detect)")
	outputDir := flag.String("output", ".", "Output directory")
	clientPath := flag.String("client", "", "Client output path (default: internal/http_client/{service})")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	specPath := flag.Arg(0)

	fmt.Printf("📖 Loading OpenAPI spec: %s\n", specPath)
	spec, err := openapi.LoadOpenAPISpec(specPath)
	if err != nil {
		log.Fatalf("Failed to load spec: %v", err)
	}

	services := openapi.DetectServices(spec)

	if len(services) == 0 {
		log.Fatalf("No services detected in OpenAPI spec")
	}

	fmt.Printf("Detected services: %v\n", services)
	fmt.Printf("Total paths: %d\n", len(spec.Paths.Map()))
	fmt.Printf("Total schemas: %d\n", len(spec.Components.Schemas))
	fmt.Println()

	var allResults []openapi.GenerationResult

	for _, svcName := range services {
		if *serviceName != "" && svcName != *serviceName {
			continue
		}

		fmt.Printf("🔨 Generating for service: %s\n", svcName)

		gen := openapi.NewGenerator(spec, svcName, "")

		result, err := gen.Generate(*outputDir, *clientPath)
		if err != nil {
			log.Fatalf("Generation failed for %s: %v", svcName, err)
		}

		fmt.Printf("Models:  %s (%d schemas)\n", result.ModelsFile, result.SchemasCount)
		fmt.Printf("Client:  %s (%d methods)\n", result.ClientFile, result.MethodsCount)
		fmt.Println()

		allResults = append(allResults, *result)
	}

	fmt.Println("Generation complete!")
	fmt.Printf("   Services generated: %d\n", len(allResults))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Review generated files")
	fmt.Println("2. Run: go build ./...")
	fmt.Println("3. Integrate methods into your tests")
}
