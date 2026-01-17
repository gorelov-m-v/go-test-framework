package openapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type Generator struct {
	spec        *openapi3.T
	serviceName string
	moduleName  string
	methods     []HTTPMethodInfo
}

type HTTPMethodInfo struct {
	Name              string
	Path              string
	HTTPMethod        string
	Operation         *openapi3.Operation
	RequestSchemaRef  string
	ResponseSchemaRef string
	PathParams        []string
}

type GenerationResult struct {
	ModelsFile   string
	ClientFile   string
	SchemasCount int
	MethodsCount int
}

func NewGenerator(spec *openapi3.T, serviceName, moduleName string) *Generator {
	return &Generator{
		spec:        spec,
		serviceName: serviceName,
		moduleName:  moduleName,
	}
}

func (g *Generator) getSanitizedName() string {
	return SanitizeServiceName(g.serviceName)
}

func LoadOpenAPISpec(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	_ = context.Background()

	return spec, nil
}

func DetectServiceName(spec *openapi3.T) string {
	if spec.Info != nil && spec.Info.Title != "" {
		name := strings.ToLower(spec.Info.Title)
		name = strings.ReplaceAll(name, " api", "")
		name = strings.ReplaceAll(name, " ", "_")
		name = strings.ReplaceAll(name, "-", "_")
		return name
	}
	return "service"
}

func DetectServices(spec *openapi3.T) []string {
	servicesSet := make(map[string]bool)

	for _, pathItem := range spec.Paths.Map() {
		for _, op := range []*openapi3.Operation{
			pathItem.Get, pathItem.Post, pathItem.Put,
			pathItem.Patch, pathItem.Delete,
		} {
			if op == nil {
				continue
			}

			if len(op.Tags) > 0 {
				serviceName := op.Tags[0]
				servicesSet[serviceName] = true
			}
		}
	}

	services := make([]string, 0, len(servicesSet))
	for name := range servicesSet {
		services = append(services, name)
	}

	if len(services) == 0 {
		services = append(services, DetectServiceName(spec))
	}

	return services
}

func (g *Generator) Generate(outputDir, clientPath string) (*GenerationResult, error) {
	sanitizedName := SanitizeServiceName(g.serviceName)

	if clientPath == "" {
		clientPath = filepath.Join(outputDir, "internal", "http_client", sanitizedName)
	}

	if err := os.MkdirAll(clientPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create client dir: %w", err)
	}

	g.collectMethods()

	modelsFile := filepath.Join(clientPath, "models.go")
	modelsCode, schemasCount, err := g.generateModels()
	if err != nil {
		return nil, fmt.Errorf("failed to generate models: %w", err)
	}
	if err := os.WriteFile(modelsFile, []byte(modelsCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write models file: %w", err)
	}

	clientFile := filepath.Join(clientPath, "client.go")
	clientCode, methodsCount, err := g.generateClient()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client: %w", err)
	}
	if err := os.WriteFile(clientFile, []byte(clientCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write client file: %w", err)
	}

	return &GenerationResult{
		ModelsFile:   modelsFile,
		ClientFile:   clientFile,
		SchemasCount: schemasCount,
		MethodsCount: methodsCount,
	}, nil
}

func getRefName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func (g *Generator) belongsToService(op *openapi3.Operation) bool {
	if len(op.Tags) == 0 {
		return false
	}

	for _, tag := range op.Tags {
		if tag == g.serviceName {
			return true
		}
	}

	return false
}

func (g *Generator) collectMethods() {
	g.methods = nil
	usedNames := make(map[string]bool)

	paths := make([]string, 0, len(g.spec.Paths.Map()))
	for path := range g.spec.Paths.Map() {
		paths = append(paths, path)
	}

	for _, path := range paths {
		pathItem := g.spec.Paths.Map()[path]

		for _, httpMethod := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
			var op *openapi3.Operation
			switch httpMethod {
			case "GET":
				op = pathItem.Get
			case "POST":
				op = pathItem.Post
			case "PUT":
				op = pathItem.Put
			case "PATCH":
				op = pathItem.Patch
			case "DELETE":
				op = pathItem.Delete
			}

			if op == nil || !g.belongsToService(op) {
				continue
			}

			methodName := g.operationToMethodName(op, path, httpMethod, usedNames)
			usedNames[methodName] = true

			info := HTTPMethodInfo{
				Name:       methodName,
				Path:       path,
				HTTPMethod: httpMethod,
				Operation:  op,
				PathParams: extractPathParams(path),
			}

			if op.RequestBody != nil && op.RequestBody.Value != nil {
				for _, content := range op.RequestBody.Value.Content {
					if content.Schema != nil && content.Schema.Ref != "" {
						info.RequestSchemaRef = content.Schema.Ref
						break
					}
				}
			}

			for _, code := range []string{"200", "201", "204"} {
				resp := op.Responses.Value(code)
				if resp == nil || resp.Value == nil {
					continue
				}
				content := resp.Value.Content.Get("application/json")
				if content != nil && content.Schema != nil && content.Schema.Ref != "" {
					info.ResponseSchemaRef = content.Schema.Ref
					break
				}
			}

			g.methods = append(g.methods, info)
		}
	}
}
