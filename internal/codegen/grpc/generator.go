package grpc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

type Generator struct {
	proto       *parser.Proto
	serviceName string
	packageName string
	moduleName  string
	pbImport    string
}

type GenerationResult struct {
	ClientFile   string
	MethodsCount int
	ServiceName  string
}

type ServiceInfo struct {
	Name    string
	Methods []MethodInfo
}

type MethodInfo struct {
	Name       string
	InputType  string
	OutputType string
	FullMethod string
}

func NewGenerator(proto *parser.Proto, serviceName, moduleName, pbImport string) *Generator {
	return &Generator{
		proto:       proto,
		serviceName: serviceName,
		moduleName:  moduleName,
		pbImport:    pbImport,
	}
}

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

func DetectServices(proto *parser.Proto) []string {
	var services []string

	for _, elem := range proto.ProtoBody {
		if service, ok := elem.(*parser.Service); ok {
			services = append(services, service.ServiceName)
		}
	}

	return services
}

func GetPackageName(proto *parser.Proto) string {
	for _, elem := range proto.ProtoBody {
		if pkg, ok := elem.(*parser.Package); ok {
			return pkg.Name
		}
	}
	return ""
}

func GetGoPackageName(proto *parser.Proto) string {
	for _, elem := range proto.ProtoBody {
		if opt, ok := elem.(*parser.Option); ok {
			if opt.OptionName == "go_package" {
				value := strings.Trim(opt.Constant, "\"")
				if idx := strings.LastIndex(value, ";"); idx != -1 {
					return value[idx+1:]
				}
				if idx := strings.LastIndex(value, "/"); idx != -1 {
					return value[idx+1:]
				}
				return value
			}
		}
	}
	return GetPackageName(proto)
}

func (g *Generator) ParseService() (*ServiceInfo, error) {
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

func (g *Generator) cleanTypeName(typeName string) string {
	typeName = strings.TrimPrefix(typeName, ".")

	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}

	return typeName
}

func (g *Generator) Generate(outputDir, clientPath string) (*GenerationResult, error) {
	sanitizedName := SanitizeServiceName(g.serviceName)

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
	clientCode, err := g.generateClient(serviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client: %w", err)
	}

	if err := os.WriteFile(clientFile, []byte(clientCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write client file: %w", err)
	}

	return &GenerationResult{
		ClientFile:   clientFile,
		MethodsCount: len(serviceInfo.Methods),
		ServiceName:  serviceInfo.Name,
	}, nil
}
