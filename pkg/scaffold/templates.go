package scaffold

import "fmt"

func (g *Generator) goModTemplate() string {
	return fmt.Sprintf(`module %s

go 1.21

require (
	github.com/gorelov-m-v/go-test-framework v0.3.0
	github.com/ozontech/allure-go/pkg/framework v0.8.0
)

// Uncomment for local development:
// replace github.com/gorelov-m-v/go-test-framework => ../go-test-framework
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
	config := `#===============================================================================
# HTTP SERVICES
#===============================================================================
http:
  jsonplaceholder:
    baseURL: "https://jsonplaceholder.typicode.com"
    timeout: 30s
    # defaultHeaders:
    #   X-Custom-Header: "value"
    # maskHeaders: "Authorization,Cookie"  # Headers to mask in Allure reports
`

	if g.options.WithDB {
		config += `
#===============================================================================
# DATABASE
#===============================================================================
database:
  mainDB:
    driver: "postgres"                    # "postgres" or "mysql"
    dsn: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
    maxOpenConns: 10
    maxIdleConns: 5
    connMaxLifetime: 1h
    # maskColumns: "password,token,secret"
`
	}

	if g.options.WithGRPC {
		config += `
#===============================================================================
# gRPC SERVICES
#===============================================================================
grpc:
  playerService:
    target: "localhost:9090"
    timeout: 30s
    insecure: true
`
	}

	if g.options.WithRedis {
		config += `
#===============================================================================
# REDIS
#===============================================================================
redis:
  cache:
    addr: "localhost:6379"
    password: ""
    db: 0
`
	}

	if g.options.WithKafka {
		config += `
#===============================================================================
# KAFKA
#===============================================================================
kafka:
  bootstrapServers: ["localhost:9092"]
  groupId: "test-group"
  topics: ["events"]
  bufferSize: 1000
`
	}

	// Async settings
	config += `
#===============================================================================
# ASYNC RETRY SETTINGS
#===============================================================================
http_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 2s
    jitter: 0.2
`

	if g.options.WithDB {
		config += `
db_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 500ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 2s
    jitter: 0.1
`
	}

	if g.options.WithGRPC {
		config += `
grpc_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 2s
    jitter: 0.2
`
	}

	if g.options.WithRedis {
		config += `
redis_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 1s
    jitter: 0.1
`
	}

	if g.options.WithKafka {
		config += `
kafka_dsl:
  async:
    enabled: true
    timeout: 30s
    interval: 500ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 3s
    jitter: 0.2
`
	}

	return config
}

func (g *Generator) envTemplate() string {
	if g.options.WithExample {
		return fmt.Sprintf(`package tests

import (
	"log"

	"%s/internal/client/jsonplaceholder"

	"github.com/gorelov-m-v/go-test-framework/pkg/builder"
)

var env *TestEnv

// TestEnv contains all dependencies for tests
type TestEnv struct {
	// HTTP clients
	JSONPlaceholder jsonplaceholder.Link `+"`config:\"jsonplaceholder\"`"+`

	// gRPC clients
	// PlayerGRPC player.Link `+"`grpc_config:\"playerService\"`"+`

	// Database repositories
	// Users users.Link `+"`db_config:\"mainDB\"`"+`

	// Redis
	// Cache cache.Link `+"`redis_config:\"cache\"`"+`

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
	// MyService myservice.Link ` + "`config:\"myService\"`" + `

	// gRPC clients
	// PlayerGRPC player.Link ` + "`grpc_config:\"playerService\"`" + `

	// Database repositories
	// Users users.Link ` + "`db_config:\"mainDB\"`" + `

	// Redis
	// Cache cache.Link ` + "`redis_config:\"cache\"`" + `

	// Kafka
	// Events events.Link ` + "`kafka_config:\"kafka\"`" + `
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

	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	"%s/internal/client/jsonplaceholder"
)

type ExampleSuite struct {
	extension.BaseSuite
}

func TestExampleSuite(t *testing.T) {
	suite.RunNamedSuite(t, "Example API Tests", new(ExampleSuite))
}

func (s *ExampleSuite) TestGetPost(t provider.T) {
	t.Title("Get post by ID")

	s.Step(t, "Get post with ID=1", func(sCtx provider.StepCtx) {
		jsonplaceholder.GetPost(sCtx, "1").
			ExpectResponseStatus(200).
			ExpectResponseBodyFieldValue("id", float64(1)).
			ExpectResponseBodyFieldNotEmpty("title").
			Send()
	})
}
`, g.moduleName)
}

func (g *Generator) exampleClientTemplate() string {
	return `package jsonplaceholder

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

// GetPost retrieves a post by ID
func GetPost(sCtx provider.StepCtx, id string) *dsl.Call[dsl.EmptyRequest, Post] {
	return dsl.NewCall[dsl.EmptyRequest, Post](sCtx, httpClient).
		GET("/posts/{id}").
		PathParam("id", id)
}
`
}

func (g *Generator) exampleModelsTemplate() string {
	return `package jsonplaceholder

// Post represents a post from JSONPlaceholder API
type Post struct {
	ID     int    ` + "`json:\"id\"`" + `
	UserID int    ` + "`json:\"userId\"`" + `
	Title  string ` + "`json:\"title\"`" + `
	Body   string ` + "`json:\"body\"`" + `
}
`
}
