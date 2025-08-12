// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"rq/dock"
	"rq/variable"
	"strconv"
	"strings"
	"time"
)

type HttpRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
	Version string
}

type HttpResponse struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       string
	Duration   time.Duration
}

func ParseHttpRequest(content string) (*HttpRequest, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty request content")
	}

	requestLine := strings.TrimSpace(lines[0])
	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid request line: %s", requestLine)
	}

	req := &HttpRequest{
		Method:  parts[0],
		URL:     parts[1],
		Headers: make(map[string]string),
		Version: "HTTP/1.1",
	}

	if len(parts) >= 3 {
		req.Version = parts[2]
	}

	i := 1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			break
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			return nil, fmt.Errorf("invalid header format: %s", line)
		}

		key := strings.TrimSpace(line[:colonIndex])
		value := strings.TrimSpace(line[colonIndex+1:])
		req.Headers[key] = value
		i++
	}

	if i < len(lines) {
		bodyLines := lines[i:]
		req.Body = strings.Join(bodyLines, "\n")
		req.Body = strings.TrimSpace(req.Body)
	}

	return req, nil
}

func (req *HttpRequest) Execute() (*HttpResponse, error) {
	start := time.Now()

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		req.URL = parsedURL.String()
	}

	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	if req.Body != "" && httpReq.Header.Get("Content-Length") == "" {
		httpReq.Header.Set("Content-Length", strconv.Itoa(len(req.Body)))
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	duration := time.Since(start)

	response := &HttpResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       string(bodyBytes),
		Duration:   duration,
	}

	return response, nil
}

// PrintResponse prints a formatted response to stdout
func (resp *HttpResponse) Print() {
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Duration: %v\n", resp.Duration)
	fmt.Println("Headers:")
	for key, values := range resp.Headers {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	fmt.Println("Body:")
	fmt.Println(resp.Body)
}

func (resp *HttpResponse) SaveToFile(filename string) error {
	content := fmt.Sprintf("Status: %s\nDuration: %v\n\nHeaders:\n", resp.Status, resp.Duration)
	for key, values := range resp.Headers {
		for _, value := range values {
			content += fmt.Sprintf("%s: %s\n", key, value)
		}
	}
	content += "\nBody:\n" + resp.Body

	return os.WriteFile(filename, []byte(content), 0644)
}

func Evaluate(ctx *dock.RqContext, request string) error {
	requestPath := filepath.Join(ctx.Dock, request)
	if !strings.HasSuffix(requestPath, ".http") {
		requestPath += ".http"
	}

	config, err := ctx.GetConfig(filepath.Dir(request))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	resolver := variable.NewVariableResolver(config)
	content, err := resolver.ResolveFile(requestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve variables in request: %w", err)
	}

	httpReq, err := ParseHttpRequest(content)
	if err != nil {
		return fmt.Errorf("failed to parse HTTP request: %w", err)
	}

	fmt.Printf("Executing %s %s\n", httpReq.Method, httpReq.URL)
	response, err := httpReq.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	response.Print()

	return nil
}

// func EvaluateWithOptions(ctx *dock.RqContext, request string, options ExecuteOptions) error {
// 	requestPath := filepath.Join(ctx.Dock, request)
// 	if !strings.HasSuffix(requestPath, ".http") {
// 		requestPath += ".http"
// 	}
//
// 	var config map[string]string
// 	var err error
// 	if options.Environment != "" {
// 		config, err = ctx.GetConfigForEnv(filepath.Dir(request), options.Environment)
// 	} else {
// 		config, err = ctx.GetConfig(filepath.Dir(request))
// 	}
// 	if err != nil {
// 		return fmt.Errorf("failed to load config: %w", err)
// 	}
//
// 	resolver := variable.NewVariableResolver(config)
// 	content, err := resolver.ResolveFile(requestPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to resolve variables in request: %w", err)
// 	}
//
// 	httpReq, err := ParseHttpRequest(content)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse HTTP request: %w", err)
// 	}
//
// 	fmt.Printf("Executing %s %s\n", httpReq.Method, httpReq.URL)
// 	response, err := httpReq.Execute()
// 	if err != nil {
// 		return fmt.Errorf("failed to execute request: %w", err)
// 	}
//
// 	if options.OutputFile != "" {
// 		if options.OutputBodyOnly {
// 			err = os.WriteFile(options.OutputFile, []byte(response.Body), 0644)
// 		} else {
// 			err = response.SaveToFile(options.OutputFile)
// 		}
// 		if err != nil {
// 			return fmt.Errorf("failed to save output: %w", err)
// 		}
// 		fmt.Printf("Response saved to %s\n", options.OutputFile)
// 	} else {
// 		response.Print()
// 	}
//
// 	return nil
// }

type ExecuteOptions struct {
	Environment    string
	OutputFile     string
	OutputBodyOnly bool
	Timeout        time.Duration
}
