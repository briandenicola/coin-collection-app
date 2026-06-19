package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type registeredRoute struct {
	Method string
	Path   string
	Line   int
}

type openAPIDocument struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

var intentionallyUndocumentedRoutes = map[string]string{
	"GET /health":            "container orchestration health check, not part of the /api contract",
	"GET /healthz":           "container orchestration health check, not part of the /api contract",
	"GET /swagger/*any":      "Swagger UI asset route, not an API endpoint",
	"GET /uploads/*filepath": "root-level authenticated media alias; /api/uploads/{filepath} is the documented API route",
}

func TestRegisteredAPIRoutesAreDocumentedInOpenAPI(t *testing.T) {
	routes, err := parseRegisteredRoutes("main.go")
	if err != nil {
		t.Fatalf("parse registered routes: %v", err)
	}
	operations, err := parseOpenAPIOperations("docs/swagger.json")
	if err != nil {
		t.Fatalf("parse OpenAPI operations: %v", err)
	}

	var missing []string
	for _, route := range routes {
		if isRouteIntentionallyUndocumented(route) {
			continue
		}
		openAPIPath := routeToOpenAPIPath(route.Path)
		key := route.Method + " " + openAPIPath
		if !operations[key] {
			missing = append(missing, fmt.Sprintf("%s %s (main.go:%d, expected OpenAPI path %s)", route.Method, route.Path, route.Line, openAPIPath))
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("registered public API routes missing from OpenAPI:\n%s\n\nIf a route is intentionally non-public/internal, document it in intentionallyUndocumentedRoutes or isRouteIntentionallyUndocumented.", strings.Join(missing, "\n"))
	}
}

func parseRegisteredRoutes(filename string) ([]registeredRoute, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	groupPrefixes := map[string]string{"r": ""}
	groupPattern := regexp.MustCompile(`^\s*(\w+)\s*:=\s*(\w+)\.Group\("([^"]*)"\)`)
	routePattern := regexp.MustCompile(`\b(\w+)\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"`)

	var routes []registeredRoute
	for lineNumber, line := range lines {
		if matches := groupPattern.FindStringSubmatch(line); matches != nil {
			group, parent, suffix := matches[1], matches[2], matches[3]
			groupPrefixes[group] = groupPrefixes[parent] + suffix
			continue
		}
		matches := routePattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		receiver, method, path := matches[1], matches[2], matches[3]
		prefix, ok := groupPrefixes[receiver]
		if !ok {
			return nil, fmt.Errorf("line %d: route receiver %q has no known group prefix", lineNumber+1, receiver)
		}
		routes = append(routes, registeredRoute{
			Method: method,
			Path:   prefix + path,
			Line:   lineNumber + 1,
		})
	}
	return routes, nil
}

func parseOpenAPIOperations(filename string) (map[string]bool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var doc openAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	operations := make(map[string]bool)
	for path, methods := range doc.Paths {
		for method := range methods {
			operations[strings.ToUpper(method)+" "+path] = true
		}
	}
	return operations, nil
}

func isRouteIntentionallyUndocumented(route registeredRoute) bool {
	if _, ok := intentionallyUndocumentedRoutes[route.Method+" "+route.Path]; ok {
		return true
	}
	return strings.HasPrefix(route.Path, "/api/internal/tools/")
}

func routeToOpenAPIPath(path string) string {
	path = strings.TrimPrefix(path, "/api")
	if path == "" {
		path = "/"
	}
	path = regexp.MustCompile(`:([A-Za-z_][A-Za-z0-9_]*)`).ReplaceAllString(path, `{$1}`)
	path = regexp.MustCompile(`\*([A-Za-z_][A-Za-z0-9_]*)`).ReplaceAllString(path, `{$1}`)
	return path
}
