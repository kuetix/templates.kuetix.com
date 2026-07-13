package transitions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/kuetix/engine/engine/domain"
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
	"gopkg.in/yaml.v3"
)

type swaggerFileTransitions struct {
	workflow.BaseServiceTransition
	initOnce sync.Once
}

func NewSwaggerFileTransitions() interfaces.ServiceTransitions {
	return &swaggerFileTransitions{}
}

// ServeSwaggerSpec serves the swagger specification files
func (s *swaggerFileTransitions) ServeSwaggerSpec() (r domain.FlowStepResult) {
	// Get absolute paths for security
	cwd, err := os.Getwd()
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to get working directory: %w", err)
		return
	}

	swaggerJSONPath := filepath.Join(cwd, "docs", "swagger.json")
	swaggerYAMLPath := filepath.Join(cwd, "docs", "swagger.yaml")

	// Verify files exist
	if _, err := os.Stat(swaggerJSONPath); os.IsNotExist(err) {
		r.Success = false
		r.Error = fmt.Errorf("swagger.json not found at %s - please ensure docs/swagger.json exists", swaggerJSONPath)
		return
	}
	if _, err := os.Stat(swaggerYAMLPath); os.IsNotExist(err) {
		r.Success = false
		r.Error = fmt.Errorf("swagger.yaml not found at %s - please ensure docs/swagger.yaml exists", swaggerYAMLPath)
		return
	}

	// Serve swagger.json
	http.HandleFunc("/docs/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, req, swaggerJSONPath)
	})

	// Serve swagger.yaml
	http.HandleFunc("/docs/swagger.yaml", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, req, swaggerYAMLPath)
	})

	r.Success = true
	r.Response = map[string]interface{}{
		"message":     "Swagger spec files registered",
		"swaggerJSON": swaggerJSONPath,
		"swaggerYAML": swaggerYAMLPath,
	}
	return
}

// RouteEndpoint represents a single endpoint definition
type RouteEndpoint struct {
	Method      string   `json:"method"`
	Workflow    string   `json:"workflow"`
	Description string   `json:"description"`
	Require     *Require `json:"require,omitempty"`
}

// Require represents the required fields for an endpoint
type Require struct {
	Headers []string `json:"headers,omitempty"`
	QS      []string `json:"qs,omitempty"`
	JSON    []string `json:"json,omitempty"`
}

// OpenAPI3 represents the OpenAPI 3.0 specification
type OpenAPI3 struct {
	OpenAPI    string              `json:"openapi"`
	Info       OpenAPIInfo         `json:"info"`
	Servers    []Server            `json:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components,omitempty"`
}

// OpenAPIInfo contains metadata about the API
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Server represents a server configuration
type Server struct {
	URL         string              `json:"url"`
	Description string              `json:"description,omitempty"`
	Variables   map[string]Variable `json:"variables,omitempty"`
}

// Variable represents a server variable
type Variable struct {
	Default string   `json:"default"`
	Enum    []string `json:"enum,omitempty"`
}

// PathItem represents a path in the API
type PathItem struct {
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
}

// Parameter represents an operation parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, header, path, cookie
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBody represents the request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType represents the media type of the request/response
type MediaType struct {
	Schema   *Schema            `json:"schema"`
	Example  interface{}        `json:"example,omitempty"`
	Examples map[string]Example `json:"examples,omitempty"`
}

// Schema represents a JSON Schema object
type Schema struct {
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
	AnyOf      []Schema          `json:"anyOf,omitempty"`
	AllOf      []Schema          `json:"allOf,omitempty"`
	OneOf      []Schema          `json:"oneOf,omitempty"`
	Enum       []interface{}     `json:"enum,omitempty"`
}

// Response represents an API response
type Response struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Header represents a response header
type Header struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema,omitempty"`
}

// Example represents an example value
type Example struct {
	Summary       string      `json:"summary,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

// SecurityRequirement represents a security requirement
type SecurityRequirement struct {
	BearerAuth []string `json:"bearerAuth,omitempty"`
}

// Components contains reusable components
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type             string      `json:"type"` // "apiKey", "http", "oauth2", "openIdConnect"
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`             // for apiKey
	In               string      `json:"in,omitempty"`               // for apiKey: "query", "header", "cookie"
	Scheme           string      `json:"scheme,omitempty"`           // for http
	BearerFormat     string      `json:"bearerFormat,omitempty"`     // for http with bearer
	Flows            OAuth2Flows `json:"flows,omitempty"`            // for oauth2
	OpenIdConnectUrl string      `json:"openIdConnectUrl,omitempty"` // for openIdConnect
}

// OAuth2Flows represents OAuth2 flows
type OAuth2Flows struct {
	Implicit          *OAuth2Flow `json:"implicit,omitempty"`
	Password          *OAuth2Flow `json:"password,omitempty"`
	ClientCredentials *OAuth2Flow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuth2Flow `json:"authorizationCode,omitempty"`
}

// OAuth2Flow represents a single OAuth2 flow
type OAuth2Flow struct {
	AuthorizationUrl string            `json:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty"`
	RefreshUrl       string            `json:"refreshUrl,omitempty"`
}

// GenerateSwagger generates OpenAPI/Swagger documentation from routes.wsl
func (s *swaggerFileTransitions) GenerateSwagger() (r domain.FlowStepResult) {
	// Get absolute paths
	cwd, err := os.Getwd()
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to get working directory: %w", err)
		return
	}

	routesPath := filepath.Join(cwd, "workflows", "api_server", "routes.wsl")
	docsDir := filepath.Join(cwd, "docs")
	swaggerJSONPath := filepath.Join(docsDir, "swagger.json")
	swaggerYAMLPath := filepath.Join(docsDir, "swagger.yaml")

	// Ensure docs directory exists
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to create docs directory: %w", err)
		return
	}

	// Read routes.wsl file
	routesContent, err := os.ReadFile(routesPath)
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to read routes.wsl: %w", err)
		return
	}

	// Parse routes from the file
	routes := s.parseRoutes(string(routesContent))

	// Generate OpenAPI specification
	openAPI := s.generateOpenAPI(routes)

	// Write swagger.json
	swaggerJSON, err := json.MarshalIndent(openAPI, "", "  ")
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to marshal swagger JSON: %w", err)
		return
	}

	if err := os.WriteFile(swaggerJSONPath, swaggerJSON, 0644); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to write swagger.json: %w", err)
		return
	}

	// Write swagger.yaml
	swaggerYAML, err := yaml.Marshal(openAPI)
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to marshal swagger YAML: %w", err)
		return
	}

	if err := os.WriteFile(swaggerYAMLPath, swaggerYAML, 0644); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to write swagger.yaml: %w", err)
		return
	}

	r.Success = true
	r.StatusCode = http.StatusOK
	r.Response = map[string]interface{}{
		"message":     "Swagger documentation generated successfully",
		"swaggerJSON": swaggerJSONPath,
		"swaggerYAML": swaggerYAMLPath,
		"endpoints":   len(routes),
	}
	return
}

// parseRoutes parses the routes.wsl file content and extracts route information
func (s *swaggerFileTransitions) parseRoutes(content string) map[string][]RouteEndpoint {
	routes := make(map[string][]RouteEndpoint)

	// Remove comments (but preserve strings)
	content = s.removeComments(content)

	// Extract route path and its endpoint array
	// Pattern: "/path": [{...}, ...],
	routeRegex := regexp.MustCompile(`"(/[^"]+)":\s*\[([^]]+)]`)
	matches := routeRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		path := match[1]
		endpointsStr := match[2]

		// Parse endpoints
		endpoints := s.parseEndpoints(endpointsStr)
		if len(endpoints) > 0 {
			routes[path] = endpoints
		}
	}

	return routes
}

// removeComments removes JavaScript-style comments from the content
func (s *swaggerFileTransitions) removeComments(content string) string {
	// Remove single-line comments
	singleLineComment := regexp.MustCompile(`//.*`)
	content = singleLineComment.ReplaceAllString(content, "")

	// Remove multi-line comments
	multiLineComment := regexp.MustCompile(`/\*.*?\*/`)
	content = multiLineComment.ReplaceAllString(content, "")

	return content
}

// parseEndpoints parses the endpoint array string into RouteEndpoint structs
func (s *swaggerFileTransitions) parseEndpoints(endpointsStr string) []RouteEndpoint {
	var endpoints []RouteEndpoint

	// Pattern to match individual endpoint objects
	endpointRegex := regexp.MustCompile(`\{[^{}]+}`)
	endpointMatches := endpointRegex.FindAllString(endpointsStr, -1)

	for _, endpointStr := range endpointMatches {
		endpoint := s.parseEndpoint(endpointStr)
		if endpoint.Method != "" {
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

// parseEndpoint parses a single endpoint object string
func (s *swaggerFileTransitions) parseEndpoint(endpointStr string) RouteEndpoint {
	var endpoint RouteEndpoint

	// Extract method
	methodRegex := regexp.MustCompile(`"method":\s*"([^"]+)"`)
	if match := methodRegex.FindStringSubmatch(endpointStr); len(match) > 1 {
		endpoint.Method = match[1]
	}

	// Extract workflow
	workflowRegex := regexp.MustCompile(`"workflow":\s*"([^"]+)"`)
	if match := workflowRegex.FindStringSubmatch(endpointStr); len(match) > 1 {
		endpoint.Workflow = match[1]
	}

	// Extract description
	descRegex := regexp.MustCompile(`"description":\s*"([^"]+)"`)
	if match := descRegex.FindStringSubmatch(endpointStr); len(match) > 1 {
		endpoint.Description = match[1]
	}

	// Extract require object
	requireRegex := regexp.MustCompile(`"require":\s*\{([^}]+)}`)
	if match := requireRegex.FindStringSubmatch(endpointStr); len(match) > 1 {
		endpoint.Require = s.parseRequire(match[1])
	}

	return endpoint
}

// parseRequire parses the require object string
func (s *swaggerFileTransitions) parseRequire(requireStr string) *Require {
	require := &Require{}

	// Extract headers array
	headersRegex := regexp.MustCompile(`"headers":\s*\[([^]]+)]`)
	if match := headersRegex.FindStringSubmatch(requireStr); len(match) > 1 {
		require.Headers = s.parseStringArray(match[1])
	}

	// Extract qs array
	qsRegex := regexp.MustCompile(`"qs":\s*\[([^]]+)]`)
	if match := qsRegex.FindStringSubmatch(requireStr); len(match) > 1 {
		require.QS = s.parseStringArray(match[1])
	}

	// Extract json array
	jsonRegex := regexp.MustCompile(`"json":\s*\[([^]]+)]`)
	if match := jsonRegex.FindStringSubmatch(requireStr); len(match) > 1 {
		require.JSON = s.parseStringArray(match[1])
	}

	return require
}

// parseStringArray parses a string array like "item1","item2","item3"
func (s *swaggerFileTransitions) parseStringArray(arrayStr string) []string {
	var result []string

	// Match quoted strings
	elementRegex := regexp.MustCompile(`"([^"]+)"`)
	matches := elementRegex.FindAllStringSubmatch(arrayStr, -1)

	for _, match := range matches {
		if len(match) > 1 {
			result = append(result, match[1])
		}
	}

	return result
}

// generateOpenAPI generates the OpenAPI 3.0 specification from parsed routes
func (s *swaggerFileTransitions) generateOpenAPI(routes map[string][]RouteEndpoint) *OpenAPI3 {
	openAPI := &OpenAPI3{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       "Kuetix API",
			Version:     "1.0.0",
			Description: "API documentation for Kuetix - A workflow-based API server",
		},
		Servers: []Server{
			{
				URL:         "/api/v1",
				Description: "Production server",
			},
			{
				URL:         "http://localhost:8080",
				Description: "Development server",
			},
		},
		Paths: make(map[string]PathItem),
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
					Description:  "Enter your JWT token",
				},
			},
			Schemas: s.generateSchemas(),
		},
	}

	// Sort paths for consistent output
	paths := make([]string, 0, len(routes))
	for path := range routes {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	// Process each route
	for _, path := range paths {
		endpoints := routes[path]
		pathItem := PathItem{}

		for _, endpoint := range endpoints {
			operation := s.generateOperation(endpoint, path)

			// Add operation to path item based on HTTP method
			switch endpoint.Method {
			case "GET":
				pathItem.Get = operation
			case "POST":
				pathItem.Post = operation
			case "PUT":
				pathItem.Put = operation
			case "DELETE":
				pathItem.Delete = operation
			case "PATCH":
				pathItem.Patch = operation
			case "OPTIONS":
				pathItem.Options = operation
			case "HEAD":
				pathItem.Head = operation
			case "TRACE":
				pathItem.Trace = operation
			}
		}

		if pathItem.Get != nil || pathItem.Post != nil || pathItem.Put != nil ||
			pathItem.Delete != nil || pathItem.Patch != nil || pathItem.Options != nil ||
			pathItem.Head != nil || pathItem.Trace != nil {
			openAPI.Paths[path] = pathItem
		}
	}

	return openAPI
}

// generateOperation generates an Operation object from a RouteEndpoint
func (s *swaggerFileTransitions) generateOperation(endpoint RouteEndpoint, path string) *Operation {
	operation := &Operation{
		Summary:     endpoint.Description,
		Description: endpoint.Description,
		OperationID: s.generateOperationID(endpoint, path),
		Responses:   make(map[string]Response),
	}

	// Add tags based on path
	operation.Tags = s.extractTags(path)

	// Add parameters from require.qs (query string parameters)
	if endpoint.Require != nil && len(endpoint.Require.QS) > 0 {
		for _, paramName := range endpoint.Require.QS {
			param := Parameter{
				Name:        paramName,
				In:          "query",
				Description: s.inferDescription(paramName),
				Required:    true,
				Schema:      s.inferSchema(paramName),
			}
			operation.Parameters = append(operation.Parameters, param)
		}
	}

	// Add header parameters from require.headers
	if endpoint.Require != nil && len(endpoint.Require.Headers) > 0 {
		for _, headerName := range endpoint.Require.Headers {
			param := Parameter{
				Name:        headerName,
				In:          "header",
				Description: s.inferDescription(headerName),
				Required:    true,
				Schema:      &Schema{Type: "string"},
			}
			operation.Parameters = append(operation.Parameters, param)

			// If it's Authorization header, add bearer auth security
			if headerName == "Authorization" {
				operation.Security = []SecurityRequirement{
					{BearerAuth: []string{}},
				}
			}
		}
	}

	// Add request body from require.json
	if endpoint.Require != nil && len(endpoint.Require.JSON) > 0 {
		operation.RequestBody = &RequestBody{
			Description: "Request body",
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: s.generateRequestBodySchema(endpoint.Require.JSON),
				},
			},
		}
	}

	// Add standard responses
	operation.Responses["200"] = Response{
		Description: "Successful response",
		Content: map[string]MediaType{
			"application/json": {
				Schema: &Schema{
					Type: "object",
					Properties: map[string]Schema{
						"success": {Type: "boolean"},
						"data":    {Type: "object"},
					},
				},
			},
		},
	}

	if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
		operation.Responses["201"] = Response{
			Description: "Resource created",
		}
	}

	operation.Responses["400"] = Response{
		Description: "Bad Request - Invalid input",
	}

	operation.Responses["401"] = Response{
		Description: "Unauthorized - Missing or invalid token",
	}

	operation.Responses["404"] = Response{
		Description: "Not Found - Resource not found",
	}

	operation.Responses["500"] = Response{
		Description: "Internal Server Error",
	}

	return operation
}

// generateOperationID generates a unique operation ID
func (s *swaggerFileTransitions) generateOperationID(endpoint RouteEndpoint, path string) string {
	// Clean path: remove leading slash, replace special chars
	cleanPath := strings.TrimPrefix(path, "/")
	cleanPath = strings.ReplaceAll(cleanPath, "/", "_")
	cleanPath = strings.ReplaceAll(cleanPath, ":", "_")

	// Add HTTP method prefix
	method := strings.ToLower(endpoint.Method)
	return fmt.Sprintf("%s_%s", method, cleanPath)
}

// extractTags extracts tags from the path
func (s *swaggerFileTransitions) extractTags(path string) []string {
	tags := []string{}

	// First path segment is usually the tag
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		tag := s.capitalizeFirst(parts[0])
		tags = append(tags, tag)
	}

	return tags
}

// capitalizeFirst capitalizes the first letter of a string
func (s *swaggerFileTransitions) capitalizeFirst(str string) string {
	if len(str) == 0 {
		return str
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

// inferDescription infers a description from a parameter name
func (s *swaggerFileTransitions) inferDescription(name string) string {
	// Convert camelCase or snake_case to readable format
	result := strings.ReplaceAll(name, "_", " ")
	result = strings.ReplaceAll(result, " ", " ")

	// Capitalize first letter
	if len(result) > 0 {
		result = strings.ToUpper(result[:1]) + result[1:]
	}

	return result
}

// inferSchema infers a schema based on parameter name
func (s *swaggerFileTransitions) inferSchema(name string) *Schema {
	// Default to string
	schema := &Schema{Type: "string"}

	// Infer type from name
	switch {
	case strings.Contains(strings.ToLower(name), "id") || strings.Contains(strings.ToLower(name), "count") ||
		strings.Contains(strings.ToLower(name), "page") || strings.Contains(strings.ToLower(name), "size") ||
		strings.Contains(strings.ToLower(name), "limit") || strings.Contains(strings.ToLower(name), "offset"):
		schema.Type = "integer"
	case strings.Contains(strings.ToLower(name), "price") || strings.Contains(strings.ToLower(name), "amount") ||
		strings.Contains(strings.ToLower(name), "cost") || strings.Contains(strings.ToLower(name), "value"):
		schema.Type = "number"
		schema.Format = "double"
	case strings.Contains(strings.ToLower(name), "active") || strings.Contains(strings.ToLower(name), "enabled") ||
		strings.Contains(strings.ToLower(name), "deleted") || strings.Contains(strings.ToLower(name), "valid") ||
		strings.Contains(strings.ToLower(name), "increment") || strings.Contains(strings.ToLower(name), "sort"):
		schema.Type = "boolean"
	}

	return schema
}

// generateRequestBodySchema generates a schema for the request body
func (s *swaggerFileTransitions) generateRequestBodySchema(fields []string) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
		Required:   make([]string, 0, len(fields)),
	}

	for _, field := range fields {
		schema.Properties[field] = *s.inferSchema(field)
		schema.Required = append(schema.Required, field)
	}

	return schema
}

// generateSchemas generates common schemas for reuse
func (s *swaggerFileTransitions) generateSchemas() map[string]Schema {
	return map[string]Schema{
		"User": {
			Type: "object",
			Properties: map[string]Schema{
				"id":          {Type: "string", Format: "uuid"},
				"email":       {Type: "string", Format: "email"},
				"username":    {Type: "string"},
				"fullName":    {Type: "string"},
				"bio":         {Type: "string"},
				"avatar":      {Type: "string", Format: "uri"},
				"location":    {Type: "string"},
				"website":     {Type: "string", Format: "uri"},
				"company":     {Type: "string"},
				"socialLinks": {Type: "array", Items: &Schema{Type: "string"}},
				"createdAt":   {Type: "string", Format: "date-time"},
				"updatedAt":   {Type: "string", Format: "date-time"},
			},
			Required: []string{"id", "email", "username", "fullName"},
		},
		"Package": {
			Type: "object",
			Properties: map[string]Schema{
				"name":         {Type: "string"},
				"version":      {Type: "string"},
				"description":  {Type: "string"},
				"author":       {Type: "string"},
				"license":      {Type: "string"},
				"dependencies": {Type: "object"},
			},
			Required: []string{"name", "version"},
		},
		"Project": {
			Type: "object",
			Properties: map[string]Schema{
				"id":           {Type: "string", Format: "uuid"},
				"name":         {Type: "string"},
				"templateName": {Type: "string"},
				"outputPath":   {Type: "string"},
				"variables":    {Type: "object"},
				"createdAt":    {Type: "string", Format: "date-time"},
			},
			Required: []string{"id", "name", "templateName"},
		},
	}
}
