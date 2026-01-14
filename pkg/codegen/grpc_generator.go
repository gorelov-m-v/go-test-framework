package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// GRPCGenerator generates DSL client code from .proto files
type GRPCGenerator struct {
	proto       *parser.Proto
	serviceName string
	packageName string
	moduleName  string
	pbImport    string // import path for generated protobuf types
}

// GRPCGenerationResult contains paths to generated files
type GRPCGenerationResult struct {
	ClientFile   string
	MethodsCount int
	ServiceName  string
}

// ServiceInfo contains parsed service information
type ServiceInfo struct {
	Name    string
	Methods []MethodInfo
}

// MethodInfo contains parsed RPC method information
type MethodInfo struct {
	Name       string
	InputType  string
	OutputType string
	FullMethod string
}

// NewGRPCGenerator creates a new gRPC DSL generator
func NewGRPCGenerator(proto *parser.Proto, serviceName, moduleName, pbImport string) *GRPCGenerator {
	return &GRPCGenerator{
		proto:       proto,
		serviceName: serviceName,
		moduleName:  moduleName,
		pbImport:    pbImport,
	}
}

// LoadProtoFile loads and parses a .proto file
func LoadProtoFile(path string) (*parser.Proto, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open proto file: %w", err)
	}
	defer reader.Close()

	proto, err := protoparser.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto file: %w", err)
	}

	return proto, nil
}

// DetectGRPCServices returns all service names found in the proto file
func DetectGRPCServices(proto *parser.Proto) []string {
	var services []string

	for _, elem := range proto.ProtoBody {
		if service, ok := elem.(*parser.Service); ok {
			services = append(services, service.ServiceName)
		}
	}

	return services
}

// GetPackageName returns the proto package name
func GetPackageName(proto *parser.Proto) string {
	for _, elem := range proto.ProtoBody {
		if pkg, ok := elem.(*parser.Package); ok {
			return pkg.Name
		}
	}
	return ""
}

// GetGoPackageName returns the go_package option value or derives from package
func GetGoPackageName(proto *parser.Proto) string {
	for _, elem := range proto.ProtoBody {
		if opt, ok := elem.(*parser.Option); ok {
			if opt.OptionName == "go_package" {
				// Remove quotes and get package name
				value := strings.Trim(opt.Constant, "\"")
				// If it contains ;, take the part after
				if idx := strings.LastIndex(value, ";"); idx != -1 {
					return value[idx+1:]
				}
				// If it contains /, take the last part
				if idx := strings.LastIndex(value, "/"); idx != -1 {
					return value[idx+1:]
				}
				return value
			}
		}
	}
	return GetPackageName(proto)
}

// ParseService extracts service information from proto
func (g *GRPCGenerator) ParseService() (*ServiceInfo, error) {
	for _, elem := range g.proto.ProtoBody {
		service, ok := elem.(*parser.Service)
		if !ok {
			continue
		}

		if service.ServiceName != g.serviceName {
			continue
		}

		info := &ServiceInfo{
			Name: service.ServiceName,
		}

		packageName := GetPackageName(g.proto)

		for _, body := range service.ServiceBody {
			rpc, ok := body.(*parser.RPC)
			if !ok {
				continue
			}

			method := MethodInfo{
				Name:       rpc.RPCName,
				InputType:  g.cleanTypeName(rpc.RPCRequest.MessageType),
				OutputType: g.cleanTypeName(rpc.RPCResponse.MessageType),
			}

			// Build full method path: /package.Service/Method
			if packageName != "" {
				method.FullMethod = fmt.Sprintf("/%s.%s/%s", packageName, service.ServiceName, rpc.RPCName)
			} else {
				method.FullMethod = fmt.Sprintf("/%s/%s", service.ServiceName, rpc.RPCName)
			}

			info.Methods = append(info.Methods, method)
		}

		return info, nil
	}

	return nil, fmt.Errorf("service %s not found in proto file", g.serviceName)
}

// cleanTypeName removes package prefix if present
func (g *GRPCGenerator) cleanTypeName(typeName string) string {
	// Remove leading dot if present
	typeName = strings.TrimPrefix(typeName, ".")

	// If contains dot, it might be package.Type - we keep just the type
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}

	return typeName
}

// Generate generates the DSL client code
func (g *GRPCGenerator) Generate(outputDir, clientPath string) (*GRPCGenerationResult, error) {
	sanitizedName := sanitizeServiceName(g.serviceName)

	if clientPath == "" {
		clientPath = filepath.Join(outputDir, "internal", "grpc_client", sanitizedName)
	}

	if err := os.MkdirAll(clientPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create client dir: %w", err)
	}

	serviceInfo, err := g.ParseService()
	if err != nil {
		return nil, err
	}

	clientFile := filepath.Join(clientPath, "client.go")
	clientCode, err := g.generateGRPCClient(serviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client: %w", err)
	}

	if err := os.WriteFile(clientFile, []byte(clientCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write client file: %w", err)
	}

	return &GRPCGenerationResult{
		ClientFile:   clientFile,
		MethodsCount: len(serviceInfo.Methods),
		ServiceName:  serviceInfo.Name,
	}, nil
}
