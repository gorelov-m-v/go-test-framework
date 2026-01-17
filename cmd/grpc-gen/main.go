package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gorelov-m-v/go-test-framework/internal/codegen/grpc"
)

const usage = `Proto to go-test-framework gRPC DSL Generator

Usage:
    grpc-gen [options] <proto-file>

Arguments:
    proto-file    Path to .proto file

Options:
    -service string    Service name to generate (default: all services)
    -output string     Output directory (default: current directory)
    -client string     Client output path (default: internal/grpc_client/{service})
    -pb-import string  Import path for generated protobuf types (required)
    -module string     Go module name for imports (default: auto-detect from go.mod)

Examples:
    # Generate from player.proto
    grpc-gen -pb-import "myproject/pkg/pb/player" player.proto

    # Specify service name and output directory
    grpc-gen -service PlayerService -pb-import "myproject/pkg/pb" -output ./generated player.proto

    # Custom paths
    grpc-gen -client internal/grpc_client/player -pb-import "myproject/pb" player.proto

Generated files:
    - internal/grpc_client/{service}/client.go

Note: This generator creates DSL wrapper methods. The protobuf types (messages)
should be generated separately using protoc with go plugins.
`

func main() {
	serviceName := flag.String("service", "", "Service name to generate (default: all)")
	outputDir := flag.String("output", ".", "Output directory")
	clientPath := flag.String("client", "", "Client output path")
	pbImport := flag.String("pb-import", "", "Import path for generated protobuf types")
	moduleName := flag.String("module", "", "Go module name (default: auto-detect)")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	protoPath := flag.Arg(0)

	if *pbImport == "" {
		fmt.Fprintln(os.Stderr, "Error: -pb-import is required")
		fmt.Fprintln(os.Stderr, "This should be the import path where your protoc-generated Go files are located")
		fmt.Fprintln(os.Stderr, "Example: -pb-import \"myproject/pkg/pb/player\"")
		os.Exit(1)
	}

	fmt.Printf("Loading proto file: %s\n", protoPath)
	proto, err := grpc.LoadProtoFile(protoPath)
	if err != nil {
		log.Fatalf("Failed to load proto file: %v", err)
	}

	services := grpc.DetectServices(proto)
	if len(services) == 0 {
		log.Fatalf("No services found in proto file")
	}

	packageName := grpc.GetPackageName(proto)
	goPackageName := grpc.GetGoPackageName(proto)

	fmt.Printf("Proto package: %s\n", packageName)
	fmt.Printf("Go package: %s\n", goPackageName)
	fmt.Printf("Detected services: %v\n", services)
	fmt.Println()

	var allResults []grpc.GenerationResult

	for _, svcName := range services {
		if *serviceName != "" && svcName != *serviceName {
			continue
		}

		fmt.Printf("Generating for service: %s\n", svcName)

		gen := grpc.NewGenerator(proto, svcName, *moduleName, *pbImport)

		result, err := gen.Generate(*outputDir, *clientPath)
		if err != nil {
			log.Fatalf("Generation failed for %s: %v", svcName, err)
		}

		fmt.Printf("  Client: %s (%d methods)\n", result.ClientFile, result.MethodsCount)
		fmt.Println()

		allResults = append(allResults, *result)
	}

	fmt.Println("Generation complete!")
	fmt.Printf("   Services generated: %d\n", len(allResults))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Ensure protobuf types are generated: protoc --go_out=. --go-grpc_out=. your.proto")
	fmt.Println("2. Review generated client files")
	fmt.Println("3. Add Link struct to your TestEnv with grpc_config tag")
	fmt.Println("4. Run: go build ./...")
}
