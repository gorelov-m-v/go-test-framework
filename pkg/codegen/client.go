package codegen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func (g *Generator) generateClient() (string, int, error) {
	var buf strings.Builder
	methodCount := 0

	sanitizedName := g.getSanitizedName()
	buf.WriteString(fmt.Sprintf("package %s\n\n", sanitizedName))

	buf.WriteString("import (\n")
	buf.WriteString("\t\"github.com/gorelov-m-v/go-test-framework/pkg/http/client\"\n")
	buf.WriteString("\t\"github.com/gorelov-m-v/go-test-framework/pkg/http/dsl\"\n")
	buf.WriteString("\t\"github.com/ozontech/allure-go/pkg/framework/provider\"\n")
	buf.WriteString(")\n\n")

	buf.WriteString("var httpClient *client.Client\n\n")
	buf.WriteString("type Link struct{}\n\n")
	buf.WriteString("func (l *Link) SetHTTP(c *client.Client) {\n")
	buf.WriteString("\thttpClient = c\n")
	buf.WriteString("}\n\n")
	buf.WriteString("func Client() *client.Client {\n")
	buf.WriteString("\treturn httpClient\n")
	buf.WriteString("}\n\n")

	paths := make([]string, 0, len(g.spec.Paths.Map()))
	for path := range g.spec.Paths.Map() {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	usedNames := make(map[string]bool)

	for _, path := range paths {
		pathItem := g.spec.Paths.Map()[path]

		for _, methodName := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
			var op *openapi3.Operation
			switch methodName {
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

			if op == nil {
				continue
			}

			if !g.belongsToService(op) {
				continue
			}

			methodCode, err := g.generateClientMethod(path, methodName, op, usedNames)
			if err != nil {
				return "", 0, fmt.Errorf("failed to generate method for %s %s: %w", methodName, path, err)
			}

			buf.WriteString(methodCode)
			buf.WriteString("\n\n")
			methodCount++
		}
	}

	return buf.String(), methodCount, nil
}

func (g *Generator) generateClientMethod(path, httpMethod string, op *openapi3.Operation, usedNames map[string]bool) (string, error) {
	var buf strings.Builder

	funcName := g.operationToFuncName(op, path, httpMethod)

	if usedNames[funcName] {
		httpPrefix := getHTTPPrefix(httpMethod)
		funcName = httpPrefix + funcName
	}

	usedNames[funcName] = true

	pathParams := extractPathParams(path)

	reqType := g.getRequestType(op)
	respType := g.getResponseType(op)

	funcParams := []string{"sCtx provider.StepCtx"}
	for _, param := range pathParams {
		funcParams = append(funcParams, fmt.Sprintf("%s string", param))
	}

	buf.WriteString(fmt.Sprintf("func %s(%s) *dsl.Call[%s, %s] {\n",
		funcName,
		strings.Join(funcParams, ", "),
		reqType,
		respType,
	))

	buf.WriteString(fmt.Sprintf("\treturn dsl.NewCall[%s, %s](sCtx, httpClient).\n",
		reqType,
		respType,
	))

	cleanPath := g.cleanPath(path)

	buf.WriteString(fmt.Sprintf("\t\t%s(\"%s\")", httpMethod, cleanPath))

	for _, param := range pathParams {
		buf.WriteString(".\n")
		buf.WriteString(fmt.Sprintf("\t\tPathParam(\"%s\", %s)", param, param))
	}

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		if _, ok := op.RequestBody.Value.Content["application/x-www-form-urlencoded"]; ok {
			buf.WriteString(".\n")
			buf.WriteString("\t\tHeader(\"Content-Type\", \"application/x-www-form-urlencoded\")")
		}
	}

	buf.WriteString("\n}")

	return buf.String(), nil
}

func (g *Generator) operationToFuncName(op *openapi3.Operation, path string, httpMethod string) string {
	pathName := g.extractNameFromPath(path, httpMethod)

	if op.OperationID != "" {
		operationName := g.cleanOperationID(op.OperationID)

		if isGenericName(operationName) || len(operationName) < 3 {
			if pathName != "" && !isGenericName(pathName) {
				return pathName
			}
		}

		if strings.Contains(path, "{") && strings.Count(path, "/") <= 2 {
			if pathName != "" && !isGenericName(pathName) {
				return pathName
			}
		}

		if pathName != "" && !isGenericName(pathName) {
			if isCRUDPath(path) || len(pathName) < len(operationName) {
				return pathName
			}
		}

		return operationName
	}

	if pathName != "" {
		return pathName
	}

	return "Request"
}

func (g *Generator) extractNameFromPath(path string, httpMethod string) string {
	parts := strings.Split(path, "/")
	var meaningfulParts []string

	for _, part := range parts {
		if part == "" || strings.HasPrefix(part, "{") {
			continue
		}
		meaningfulParts = append(meaningfulParts, part)
	}

	if len(meaningfulParts) == 0 {
		return ""
	}

	lastPart := meaningfulParts[len(meaningfulParts)-1]

	lastPart = strings.ReplaceAll(lastPart, "-", "_")

	if len(meaningfulParts) == 1 && strings.Contains(path, "{") {
		resourceName := lastPart
		methodPrefix := httpMethodToPrefix(httpMethod)
		if methodPrefix != "" {
			return methodPrefix + snakeToCamel(resourceName)
		}
		return snakeToCamel(resourceName)
	}

	if isShortName(lastPart) {
		methodPrefix := httpMethodToPrefix(httpMethod)
		if methodPrefix != "" {
			return methodPrefix + snakeToCamel(lastPart)
		}
	}

	return snakeToCamel(lastPart)
}

func httpMethodToPrefix(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "Get"
	case "POST":
		return "Create"
	case "PUT":
		return "Update"
	case "PATCH":
		return "Update"
	case "DELETE":
		return "Delete"
	default:
		return ""
	}
}

func isShortName(name string) bool {
	// Names with 2 or fewer characters need more context
	return len(name) <= 2
}

func isGenericName(name string) bool {
	generic := map[string]bool{
		"Index": true,
		"Get":   true,
		"Post":  true,
		"Put":   true,
	}
	return generic[name]
}

func isCRUDPath(path string) bool {
	return strings.Contains(path, "/change_password") ||
		strings.Contains(path, "/forgot-password") ||
		strings.Contains(path, "/reset-password") ||
		strings.Contains(path, "/verify")
}

func (g *Generator) cleanOperationID(operationID string) string {
	httpMethods := []string{"_post", "_get", "_put", "_patch", "_delete"}
	for _, method := range httpMethods {
		operationID = strings.TrimSuffix(operationID, method)
	}

	re := regexp.MustCompile(`__[a-z_]+__`)
	operationID = re.ReplaceAllString(operationID, "_")

	words := strings.Split(operationID, "_")

	var cleanWords []string
	for _, word := range words {
		if word != "" {
			cleanWords = append(cleanWords, word)
		}
	}

	namespaceWords := map[string]bool{
		"auth":  true,
		"jwt":   true,
		"users": true,
	}

	pathParamSuffixes := map[string]bool{
		"id":      true,
		"url":     true,
		"uuid":    true,
		"userurl": true,
	}

	wordCount := make(map[string]int)
	for _, word := range cleanWords {
		wordCount[strings.ToLower(word)]++
	}

	var meaningfulWords []string
	seen := make(map[string]bool)

	for _, word := range cleanWords {
		wordLower := strings.ToLower(word)

		if seen[wordLower] {
			continue
		}

		seen[wordLower] = true

		isNamespace := namespaceWords[wordLower]
		isPathParam := pathParamSuffixes[wordLower]

		if isNamespace || isPathParam {
			continue
		}

		meaningfulWords = append(meaningfulWords, word)
	}

	if len(meaningfulWords) == 0 {
		seen = make(map[string]bool)
		for _, word := range cleanWords {
			wordLower := strings.ToLower(word)
			if !seen[wordLower] {
				meaningfulWords = append(meaningfulWords, word)
				seen[wordLower] = true
			}
		}
	}

	finalWords := meaningfulWords

	var result strings.Builder
	for _, word := range finalWords {
		if word != "" {
			result.WriteString(strings.ToUpper(word[:1]))
			if len(word) > 1 {
				result.WriteString(word[1:])
			}
		}
	}

	return result.String()
}

func (g *Generator) getRequestType(op *openapi3.Operation) string {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return "dsl.EmptyRequest"
	}

	for _, content := range op.RequestBody.Value.Content {
		if content.Schema != nil && content.Schema.Ref != "" {
			refName := getRefName(content.Schema.Ref)
			return snakeToCamel(refName)
		}
	}

	return "dsl.EmptyRequest"
}

func (g *Generator) getResponseType(op *openapi3.Operation) string {
	for _, code := range []string{"200", "201", "204"} {
		resp := op.Responses.Value(code)
		if resp == nil || resp.Value == nil {
			continue
		}

		content := resp.Value.Content.Get("application/json")
		if content != nil && content.Schema != nil && content.Schema.Ref != "" {
			refName := getRefName(content.Schema.Ref)
			return snakeToCamel(refName)
		}
	}

	return "dsl.EmptyResponse"
}

func (g *Generator) cleanPath(path string) string {
	prefixes := []string{
		"/api/v1",
		"/api",
		"/v1",
		"/" + g.getSanitizedName(),
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			break
		}
	}

	if path == "" {
		path = "/"
	}

	return path
}

func extractPathParams(path string) []string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)

	params := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}

	return params
}

func (g *Generator) getModuleName() string {
	if g.moduleName != "" {
		return g.moduleName
	}

	// TODO: implement go.mod parsing
	return "generated"
}

func getHTTPPrefix(httpMethod string) string {
	switch httpMethod {
	case "GET":
		return "Get"
	case "POST":
		return "Create"
	case "PUT":
		return "Update"
	case "PATCH":
		return "Patch"
	case "DELETE":
		return "Delete"
	default:
		return httpMethod
	}
}
