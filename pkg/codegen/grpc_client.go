package codegen

import (
	"fmt"
	"strings"
)

// generateGRPCClient generates the DSL client code for a gRPC service
func (g *GRPCGenerator) generateGRPCClient(service *ServiceInfo) (string, error) {
	var buf strings.Builder

	sanitizedName := sanitizeServiceName(g.serviceName)

	// Package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", sanitizedName))

	// Imports
	buf.WriteString("import (\n")
	buf.WriteString("\t\"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client\"\n")
	buf.WriteString("\t\"github.com/gorelov-m-v/go-test-framework/pkg/grpc/dsl\"\n")
	buf.WriteString("\t\"github.com/ozontech/allure-go/pkg/framework/provider\"\n")
	if g.pbImport != "" {
		buf.WriteString(fmt.Sprintf("\n\tpb \"%s\"\n", g.pbImport))
	}
	buf.WriteString(")\n\n")

	// Client variable and Link struct
	buf.WriteString("// grpcClient holds the gRPC client instance\n")
	buf.WriteString("var grpcClient *client.Client\n\n")

	buf.WriteString("// Link is used for auto-wiring via BuildEnv\n")
	buf.WriteString("type Link struct{}\n\n")

	buf.WriteString("// SetGRPC implements grpcclient.GRPCSetter interface\n")
	buf.WriteString("func (l *Link) SetGRPC(c *client.Client) {\n")
	buf.WriteString("\tgrpcClient = c\n")
	buf.WriteString("}\n\n")

	buf.WriteString("// Client returns the underlying gRPC client for advanced usage\n")
	buf.WriteString("func Client() *client.Client {\n")
	buf.WriteString("\treturn grpcClient\n")
	buf.WriteString("}\n")

	// Generate methods
	for _, method := range service.Methods {
		buf.WriteString("\n")
		methodCode := g.generateMethod(method)
		buf.WriteString(methodCode)
	}

	return buf.String(), nil
}

// generateMethod generates a single DSL method for an RPC
func (g *GRPCGenerator) generateMethod(method MethodInfo) string {
	var buf strings.Builder

	// Add comment
	buf.WriteString(fmt.Sprintf("// %s calls %s\n", method.Name, method.FullMethod))

	// Determine type prefix
	typePrefix := "pb."
	if g.pbImport == "" {
		typePrefix = ""
	}

	// Function signature
	buf.WriteString(fmt.Sprintf("func %s(sCtx provider.StepCtx) *dsl.Call[%s%s, %s%s] {\n",
		method.Name,
		typePrefix, method.InputType,
		typePrefix, method.OutputType,
	))

	// Function body
	buf.WriteString(fmt.Sprintf("\treturn dsl.NewCall[%s%s, %s%s](sCtx, grpcClient).\n",
		typePrefix, method.InputType,
		typePrefix, method.OutputType,
	))
	buf.WriteString(fmt.Sprintf("\t\tMethod(\"%s\")\n", method.FullMethod))
	buf.WriteString("}\n")

	return buf.String()
}
