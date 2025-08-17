// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"fmt"
	"os"
	"path/filepath"
	"rq/dock"
	"rq/variable"
	"strings"
)

func New(ctx *dock.RqContext, file string, protocol string) error {
	if file == "" {
		return fmt.Errorf("request name cannot be empty")
	}

	if protocol == "" {
		protocol = "http"
	}

	validProtocols := map[string]bool{
		"http":      true,
		"ws":        true,
		"tcp":       true,
		"websocket": true,
		"grpc":      true,
	}

	if !validProtocols[protocol] {
		return fmt.Errorf("unsupported protocol: %s (supported: http, tcp, ftp)", protocol)
	}

	dir := filepath.Dir(file)
	if dir != "." && dir != "" {
		fullDir := filepath.Join(ctx.Path, dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory %s: %w", dir, err)
		}
	}

	filename := filepath.Base(file)
	fullPath := filepath.Join(ctx.Path, file+"."+protocol)

	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("request file already exists: %s.%s", file, protocol)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}
	defer f.Close()

	template := getRequestTemplate(protocol, filename)
	if _, err := f.WriteString(template); err != nil {
		return fmt.Errorf("failed to write request template: %w", err)
	}

	return nil
}

func getRequestTemplate(protocol, name string) string {
	switch protocol {
	case "http":
		return fmt.Sprintf(`GET {{BASE_URL}}/api/%s {{HTTP_VERSION}}
User-Agent: {{USER_AGENT}}
Accept: application/json
Authorization: Bearer {{API_TOKEN}}

`, name)

	case "ws", "websocket":
		return fmt.Sprintf(`# WebSocket connection to {{BASE_URL}}
# Protocol: WebSocket
# Endpoint: /ws/%s
# 
# Connection headers:
Origin: {{BASE_URL}}
Sec-WebSocket-Protocol: json

# Messages to send:
{"type": "subscribe", "channel": "%s"}
{"type": "ping"}
`, name, name)

	case "grpc":
		return fmt.Sprintf(`# gRPC service call
# Service: {{SERVICE_NAME}}
# Method: %s
# 
# Request:
{
  "id": "{{uuid()}}",
  "timestamp": "{{timestamp()}}"
}
`, strings.Title(name))

	case "tcp":
		return fmt.Sprintf(`# The first line is the url of connection and the rest is the byte to send
{{BASE_URL}}`)
	default:
		return fmt.Sprintf(`# %s request template
# Edit this file to customize your %s request
`, strings.ToUpper(protocol), protocol)
	}
}

func Run(ctx *dock.RqContext, path string) {
	fmt.Printf("Searching for request: %s\n", path)

	requests, err := retrieveRequests(ctx.Path, path)
	if err != nil {
		fmt.Printf("Error searching for requests: %v\n", err)
		os.Exit(1)
	}

	switch len(requests) {
	case 0:
		fmt.Printf("Request '%s' not found\n", path)
		fmt.Println("Available requests:")
		showAvailableRequests(ctx.Path)
		os.Exit(1)

	case 1:
		fmt.Printf("Executing request: %s\n", requests[0])
		if err := Evaluate(ctx, path); err != nil {
			fmt.Printf("Execution failed: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Multiple requests found matching '%s':\n", path)
		for i, req := range requests {
			fmt.Printf("  %d. %s\n", i+1, req)
		}
		fmt.Println("Please be more specific.")
		os.Exit(1)
	}
}

func showAvailableRequests(basePath string) {
	requests := findAllRequests(basePath)
	if len(requests) == 0 {
		fmt.Println("  No requests found in current dock")
		fmt.Println("  Run 'rq new <name>' to create a new request")
		return
	}

	for _, req := range requests {
		relPath, _ := filepath.Rel(basePath, req)
		name := strings.TrimSuffix(relPath, filepath.Ext(relPath))
		fmt.Printf("  %s\n", name)
	}
}

func findAllRequests(basePath string) []string {
	var requests []string

	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".http" || ext == ".tcp" {
				requests = append(requests, path)
			}
		}

		return nil
	})

	return requests
}

func retrieveRequests(basePath string, reqPath string) ([]string, error) {
	var results []string

	exactPath := filepath.Join(basePath, reqPath)

	extensions := []string{".http", ".tcp"}
	for _, ext := range extensions {
		fullPath := exactPath + ext
		if _, err := os.Stat(fullPath); err == nil {
			results = append(results, fullPath)
		}
	}

	if len(results) > 0 {
		return results, nil
	}

	_, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	pathSegments := strings.Split(strings.Trim(reqPath, string(os.PathSeparator)), string(os.PathSeparator))
	currentPath := basePath

	for i, segment := range pathSegments {
		found := false

		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", currentPath, err)
		}

		for _, entry := range entries {
			if i == len(pathSegments)-1 {
				if !entry.IsDir() {
					name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
					if strings.HasPrefix(name, segment) {
						fullPath := filepath.Join(currentPath, entry.Name())
						results = append(results, fullPath)
						found = true
					}
				}
			} else {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), segment) {
					currentPath = filepath.Join(currentPath, entry.Name())
					found = true
					break
				}
			}
		}

		if !found && i < len(pathSegments)-1 {
			break
		}
	}

	return results, nil
}

func Evaluate(ctx *dock.RqContext, request string) error {
	requestPath := resolveRequestPath(ctx.Dock, request)
	if requestPath == "" {
		return fmt.Errorf("request file not found: %s", request)
	}

	config, err := ctx.GetConfig(filepath.Dir(request))
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	setDefaultVariables(config)

	resolver := variable.NewVariableResolver(config)
	content, err := resolver.ResolveFile(requestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve variables: %w", err)
	}

	ext := filepath.Ext(requestPath)
	switch ext {
	case ".http":
		return executeHTTPRequest(content)
	case ".tcp":
		return executeTCPRequest(content)
	case ".grpc":
		return fmt.Errorf("gRPC requests not yet implemented")
	default:
		return fmt.Errorf("unsupported request type: %s", ext)
	}
}

func EvaluateWithOptions(ctx *dock.RqContext, request string, options ExecuteOptions) error {
	requestPath := resolveRequestPath(ctx.Dock, request)
	if requestPath == "" {
		return fmt.Errorf("request file not found: %s", request)
	}

	var config map[string]string
	var err error

	if options.Environment != "" {
		config, err = ctx.GetConfigForEnv(filepath.Dir(request), options.Environment)
	} else {
		config, err = ctx.GetConfig(filepath.Dir(request))
	}

	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	setDefaultVariables(config)

	resolver := variable.NewVariableResolver(config)
	content, err := resolver.ResolveFile(requestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve variables: %w", err)
	}

	ext := filepath.Ext(requestPath)
	switch ext {
	case ".http":
		return executeHTTPRequestWithOptions(content, options)
	case ".ws":
		return fmt.Errorf("WebSocket requests not yet implemented")
	case ".grpc":
		return fmt.Errorf("gRPC requests not yet implemented")
	default:
		return fmt.Errorf("unsupported request type: %s", ext)
	}
}

func resolveRequestPath(dockPath, request string) string {
	extensions := []string{".http", ".ws", ".grpc"}

	basePath := filepath.Join(dockPath, request)

	for _, ext := range extensions {
		fullPath := basePath + ext
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	if filepath.Ext(request) != "" {
		fullPath := filepath.Join(dockPath, request)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}

func setDefaultVariables(config map[string]string) {
	defaults := map[string]string{
		"HTTP_VERSION": "HTTP/1.1",
		"USER_AGENT":   "rq/1.0.0",
		"ACCEPT":       "application/json",
	}

	for key, value := range defaults {
		if _, exists := config[key]; !exists {
			config[key] = value
		}
	}
}
