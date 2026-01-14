package scaffold

import "fmt"

func (g *Generator) goModTemplate() string {
	return fmt.Sprintf(`module %s

go 1.21

require (
	github.com/gorelov-m-v/go-test-framework v0.1.0
	github.com/ozontech/allure-go/pkg/framework v0.6.32
)
`, g.moduleName)
}

func gitignoreTemplate() string {
	return `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo

# Allure results
allure-results/
allure-report/

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
*.local.yaml
!config.local.yaml

# Vendor (if not committing)
# vendor/
`
}

func makefileTemplate() string {
	return `.PHONY: test test-run test-parallel allure allure-clean clean deps build lint install-tools gen-http gen-grpc gen-proto help

# Default target
.DEFAULT_GOAL := help

#===============================================================================
# TESTING
#===============================================================================

## Run all tests
test:
	go test ./tests/... -v

## Run specific test: make test-run TEST=TestName
test-run:
	go test ./tests/... -v -run $(TEST)

## Run tests in parallel (faster, but Allure report may be less readable)
test-parallel:
	go test ./tests/... -v -parallel 4

#===============================================================================
# ALLURE REPORTS
#===============================================================================

## Generate and open Allure report in browser
allure:
	allure serve allure-results

## Generate Allure report to allure-report/ directory
allure-generate:
	allure generate allure-results -o allure-report --clean

## Clean Allure results
allure-clean:
	rm -rf allure-results/*
	rm -rf allure-report/*

#===============================================================================
# CODE GENERATION
#===============================================================================

## Install code generators
install-tools:
	go install github.com/gorelov-m-v/go-test-framework/cmd/openapi-gen@latest
	go install github.com/gorelov-m-v/go-test-framework/cmd/grpc-gen@latest
	go install github.com/gorelov-m-v/go-test-framework/cmd/test-init@latest

## Generate HTTP clients from OpenAPI spec: make gen-http SPEC=openapi.json
gen-http:
	openapi-gen $(SPEC)

## Generate gRPC clients from proto: make gen-grpc PROTO=proto/service.proto PB_IMPORT=module/pb
gen-grpc:
	grpc-gen -pb-import "$(PB_IMPORT)" $(PROTO)

## Generate protobuf Go files: make gen-proto
gen-proto:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

#===============================================================================
# PROJECT MANAGEMENT
#===============================================================================

## Install/update Go dependencies
deps:
	go mod tidy
	go mod download

## Build check (verify code compiles)
build:
	go build ./...

## Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

## Clean all generated files and results
clean:
	rm -rf allure-results/*
	rm -rf allure-report/*

## Format code
fmt:
	go fmt ./...

#===============================================================================
# HELP
#===============================================================================

## Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Examples:"
	@echo "  make test                              # Run all tests"
	@echo "  make test-run TEST=TestCreatePlayer    # Run specific test"
	@echo "  make allure                            # Open Allure report"
	@echo "  make gen-http SPEC=openapi.json        # Generate HTTP clients"
	@echo "  make gen-grpc PROTO=proto/player.proto PB_IMPORT=mymodule/pb"
`
}

func (g *Generator) configTemplate() string {
	return `# HTTP services configuration
http:
  exampleService:
    baseURL: "http://localhost:8080"
    timeout: 30s
    # maskHeaders: "Authorization,Cookie"  # Headers to mask in Allure reports

# Database configuration
# database:
#   mainDB:
#     driver: "postgres"  # or "mysql"
#     dsn: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
#     # maskColumns: "password,token"  # Columns to mask in Allure reports

# Redis configuration
# redis:
#   cache:
#     addr: "localhost:6379"
#     password: ""
#     db: 0

# gRPC configuration
# grpc:
#   playerService:
#     target: "localhost:9090"
#     insecure: true

# Kafka configuration
# kafka:
#   bootstrapServers: ["localhost:9092"]
#   groupId: "test-group"
#   topics: ["events"]

# Async retry settings
http_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 2s

db_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 500ms

kafka_dsl:
  async:
    enabled: true
    timeout: 30s
    interval: 500ms

redis_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
`
}

func (g *Generator) envTemplate() string {
	if g.withExample {
		return fmt.Sprintf(`package tests

import (
	"log"

	"%s/internal/http_client/example"

	"github.com/gorelov-m-v/go-test-framework/pkg/builder"
)

var env *TestEnv

// TestEnv contains all dependencies for tests
type TestEnv struct {
	// HTTP clients
	Example example.Link `+"`config:\"exampleService\"`"+`

	// gRPC clients
	// PlayerGRPC player.Link `+"`grpc_config:\"playerService\"`"+`

	// Database repositories
	// Users users.Link `+"`db_config:\"mainDB\"`"+`

	// Redis
	// Session session.Link `+"`redis_config:\"cache\"`"+`

	// Kafka
	// Events events.Link `+"`kafka_config:\"kafka\"`"+`
}

func init() {
	env = &TestEnv{}
	if err := builder.BuildEnv(env); err != nil {
		log.Fatalf("Failed to build test environment: %%v", err)
	}
}
`, g.moduleName)
	}

	return fmt.Sprintf(`package tests

import (
	"log"

	"github.com/gorelov-m-v/go-test-framework/pkg/builder"
)

var env *TestEnv

// TestEnv contains all dependencies for tests
type TestEnv struct {
	// HTTP clients
	// Auth auth.Link `+"`config:\"authService\"`"+`

	// gRPC clients
	// PlayerGRPC player.Link `+"`grpc_config:\"playerService\"`"+`

	// Database repositories
	// Users users.Link `+"`db_config:\"mainDB\"`"+`

	// Redis
	// Session session.Link `+"`redis_config:\"cache\"`"+`

	// Kafka
	// Events events.Link `+"`kafka_config:\"kafka\"`"+`
}

func init() {
	env = &TestEnv{}
	if err := builder.BuildEnv(env); err != nil {
		log.Fatalf("Failed to build test environment: %%v", err)
	}
}
`)
}

func (g *Generator) exampleTestTemplate() string {
	return fmt.Sprintf(`package tests

import (
	"testing"

	"%s/internal/http_client/example"

	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type ExampleSuite struct {
	extension.BaseSuite
}

func TestExampleSuite(t *testing.T) {
	suite.RunNamedSuite(t, "Example API Tests", new(ExampleSuite))
}

func (s *ExampleSuite) TestHealthCheck(t provider.T) {
	t.Title("Health check endpoint")
	t.Description("Verify that the API health endpoint returns OK status")

	s.Step(t, "Check health endpoint", func(sCtx provider.StepCtx) {
		example.Health(sCtx).
			ExpectResponseStatus(200).
			ExpectResponseBodyFieldValue("status", "ok").
			Send()
	})
}

func (s *ExampleSuite) TestCreateResource(t provider.T) {
	t.Title("Create resource")
	t.Description("Create a new resource via API")

	var resp *example.CreateResourceResponse

	s.Step(t, "Create resource", func(sCtx provider.StepCtx) {
		result := example.CreateResource(sCtx).
			RequestBody(example.CreateResourceRequest{
				Name:        "test-resource",
				Description: "Test description",
			}).
			ExpectResponseStatus(201).
			ExpectResponseBodyFieldNotEmpty("id").
			Send()
		resp = &result.Body
	})

	s.Step(t, "Verify resource was created", func(sCtx provider.StepCtx) {
		sCtx.Logf("Created resource with ID: %%s", resp.ID)
	})
}
`, g.moduleName)
}

func (g *Generator) exampleClientTemplate() string {
	return `package example

import (
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/dsl"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

var httpClient *client.Client

// Link is used for auto-wiring via BuildEnv
type Link struct{}

// SetHTTP implements the HTTPSetter interface
func (l *Link) SetHTTP(c *client.Client) {
	httpClient = c
}

// Client returns the underlying HTTP client
func Client() *client.Client {
	return httpClient
}

// Health checks the API health endpoint
func Health(sCtx provider.StepCtx) *dsl.Call[dsl.EmptyRequest, HealthResponse] {
	return dsl.NewCall[dsl.EmptyRequest, HealthResponse](sCtx, httpClient).
		GET("/health")
}

// CreateResource creates a new resource
func CreateResource(sCtx provider.StepCtx) *dsl.Call[CreateResourceRequest, CreateResourceResponse] {
	return dsl.NewCall[CreateResourceRequest, CreateResourceResponse](sCtx, httpClient).
		POST("/resources")
}

// GetResource retrieves a resource by ID
func GetResource(sCtx provider.StepCtx, id string) *dsl.Call[dsl.EmptyRequest, GetResourceResponse] {
	return dsl.NewCall[dsl.EmptyRequest, GetResourceResponse](sCtx, httpClient).
		GET("/resources/{id}").
		PathParam("id", id)
}

// DeleteResource deletes a resource by ID
func DeleteResource(sCtx provider.StepCtx, id string) *dsl.Call[dsl.EmptyRequest, dsl.EmptyResponse] {
	return dsl.NewCall[dsl.EmptyRequest, dsl.EmptyResponse](sCtx, httpClient).
		DELETE("/resources/{id}").
		PathParam("id", id)
}
`
}

func (g *Generator) exampleModelsTemplate() string {
	return `package example

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string ` + "`json:\"status\"`" + `
	Version string ` + "`json:\"version,omitempty\"`" + `
}

// CreateResourceRequest represents the request to create a resource
type CreateResourceRequest struct {
	Name        string ` + "`json:\"name\"`" + `
	Description string ` + "`json:\"description,omitempty\"`" + `
}

// CreateResourceResponse represents the response after creating a resource
type CreateResourceResponse struct {
	ID          string ` + "`json:\"id\"`" + `
	Name        string ` + "`json:\"name\"`" + `
	Description string ` + "`json:\"description\"`" + `
	CreatedAt   string ` + "`json:\"created_at\"`" + `
}

// GetResourceResponse represents the response when getting a resource
type GetResourceResponse struct {
	ID          string ` + "`json:\"id\"`" + `
	Name        string ` + "`json:\"name\"`" + `
	Description string ` + "`json:\"description\"`" + `
	CreatedAt   string ` + "`json:\"created_at\"`" + `
	UpdatedAt   string ` + "`json:\"updated_at,omitempty\"`" + `
}
`
}
