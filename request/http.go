// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
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
	Timeout time.Duration
}

type HttpResponse struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       string
	Duration   time.Duration
	Size       int64
}
type ExecuteOptions struct {
	Environment    string
	OutputFile     string
	OutputBodyOnly bool
	Timeout        time.Duration
}

func ParseHttpRequest(content string) (*HttpRequest, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty request content")
	}

	lines := strings.Split(content, "\n")

	requestLine := strings.TrimSpace(lines[0])
	if requestLine == "" {
		return nil, fmt.Errorf("missing request line")
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid request line format: %s", requestLine)
	}

	req := &HttpRequest{
		Method:  strings.ToUpper(parts[0]),
		URL:     parts[1],
		Headers: make(map[string]string),
		Version: "HTTP/1.1",
		Timeout: 30 * time.Second,
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

		if strings.HasPrefix(line, "#") {
			i++
			continue
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			return nil, fmt.Errorf("invalid header format at line %d: %s", i+1, line)
		}

		key := strings.TrimSpace(line[:colonIndex])
		value := strings.TrimSpace(line[colonIndex+1:])

		if key == "" {
			return nil, fmt.Errorf("empty header name at line %d", i+1)
		}

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

	if err := req.prepareURL(); err != nil {
		return nil, fmt.Errorf("URL preparation failed: %w", err)
	}

	httpReq, err := req.createHTTPRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	client := req.createHTTPClient()

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, req.formatNetworkError(err)
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
		Size:       int64(len(bodyBytes)),
	}

	return response, nil
}

func (req *HttpRequest) prepareURL() error {
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		if strings.Contains(req.URL, "localhost") || strings.Contains(req.URL, "127.0.0.1") {
			parsedURL.Scheme = "http"
		} else {
			parsedURL.Scheme = "https"
		}
		req.URL = parsedURL.String()
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	return nil
}

func (req *HttpRequest) createHTTPRequest() (*http.Request, error) {
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	if req.Body != "" && httpReq.Header.Get("Content-Length") == "" {
		httpReq.Header.Set("Content-Length", strconv.Itoa(len(req.Body)))
	}

	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", "rq/1.0.0")
	}

	return httpReq, nil
}

func (req *HttpRequest) createHTTPClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       10,
		IdleConnTimeout:       90 * time.Second,

		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	return &http.Client{
		Timeout:   req.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
}

func (req *HttpRequest) formatNetworkError(err error) error {
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return fmt.Errorf("request timeout after %v", req.Timeout)
		}
	}

	if strings.Contains(err.Error(), "connection refused") {
		return fmt.Errorf("connection refused - server may be down or unreachable")
	}

	if strings.Contains(err.Error(), "no such host") {
		return fmt.Errorf("host not found - check the URL")
	}

	if strings.Contains(err.Error(), "certificate") {
		return fmt.Errorf("SSL/TLS certificate error: %w", err)
	}

	return fmt.Errorf("network error: %w", err)
}

func (resp *HttpResponse) Print() {
	statusColor := getStatusColor(resp.StatusCode)
	fmt.Printf("Status: %s%s%s\n", statusColor, resp.Status, "\033[0m")

	fmt.Printf("Duration: %v\n", resp.Duration)
	fmt.Printf("Size: %s\n", formatBytes(resp.Size))

	fmt.Println("\nHeaders:")
	for key, values := range resp.Headers {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println("\nBody:")
	if resp.Body == "" {
		fmt.Println("  (empty)")
	} else {
		if contentType := resp.Headers["Content-Type"]; len(contentType) > 0 && strings.Contains(contentType[0], "json") {
			if formatted := formatJSON(resp.Body); formatted != "" {
				fmt.Println(formatted)
				return
			}
		}
		fmt.Println(resp.Body)
	}
}

func (resp *HttpResponse) SaveToFile(filename string) error {
	content := resp.formatForFile()
	return os.WriteFile(filename, []byte(content), 0644)
}

func (resp *HttpResponse) formatForFile() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Status: %s\n", resp.Status))
	sb.WriteString(fmt.Sprintf("Duration: %v\n", resp.Duration))
	sb.WriteString(fmt.Sprintf("Size: %s\n", formatBytes(resp.Size)))
	sb.WriteString("\nHeaders:\n")

	for key, values := range resp.Headers {
		for _, value := range values {
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}

	sb.WriteString("\nBody:\n")
	sb.WriteString(resp.Body)

	return sb.String()
}

func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "\033[32m"
	case statusCode >= 300 && statusCode < 400:
		return "\033[33m"
	case statusCode >= 400 && statusCode < 500:
		return "\033[31m"
	case statusCode >= 500:
		return "\033[35m"
	default:
		return "\033[0m"
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatJSON(jsonStr string) string {
	var formatted strings.Builder
	var indent int
	var inString bool
	var escaped bool

	for _, char := range jsonStr {
		if escaped {
			formatted.WriteRune(char)
			escaped = false
			continue
		}

		if char == '\\' && inString {
			formatted.WriteRune(char)
			escaped = true
			continue
		}

		if char == '"' {
			inString = !inString
			formatted.WriteRune(char)
			continue
		}

		if inString {
			formatted.WriteRune(char)
			continue
		}

		switch char {
		case '{', '[':
			formatted.WriteRune(char)
			indent++
			formatted.WriteRune('\n')
			writeIndent(&formatted, indent)

		case '}', ']':
			formatted.WriteRune('\n')
			indent--
			writeIndent(&formatted, indent)
			formatted.WriteRune(char)

		case ',':
			formatted.WriteRune(char)
			formatted.WriteRune('\n')
			writeIndent(&formatted, indent)

		case ':':
			formatted.WriteRune(char)
			formatted.WriteRune(' ')

		case ' ', '\t', '\n', '\r':
			// Skip whitespace outside strings

		default:
			formatted.WriteRune(char)
		}
	}

	return formatted.String()
}

func writeIndent(sb *strings.Builder, level int) {
	for i := 0; i < level*2; i++ {
		sb.WriteRune(' ')
	}
}

func executeHTTPRequest(content string) error {
	httpReq, err := ParseHttpRequest(content)
	if err != nil {
		return fmt.Errorf("failed to parse HTTP request: %w", err)
	}

	if err := validateHTTPRequest(httpReq); err != nil {
		return fmt.Errorf("invalid HTTP request: %w", err)
	}

	fmt.Printf("Executing %s %s\n", httpReq.Method, httpReq.URL)

	response, err := httpReq.Execute()
	if err != nil {
		return fmt.Errorf("request execution failed: %w", err)
	}

	response.Print()
	return nil
}

func executeHTTPRequestWithOptions(content string, options ExecuteOptions) error {
	httpReq, err := ParseHttpRequest(content)
	if err != nil {
		return fmt.Errorf("failed to parse HTTP request: %w", err)
	}

	if err := validateHTTPRequest(httpReq); err != nil {
		return fmt.Errorf("invalid HTTP request: %w", err)
	}

	if options.Timeout > 0 {
		httpReq.Timeout = options.Timeout
	}

	fmt.Printf("Executing %s %s", httpReq.Method, httpReq.URL)
	if options.Environment != "" {
		fmt.Printf(" (env: %s)", options.Environment)
	}
	fmt.Println()

	response, err := httpReq.Execute()
	if err != nil {
		return fmt.Errorf("request execution failed: %w", err)
	}

	if options.OutputFile != "" {
		if options.OutputBodyOnly {
			err = os.WriteFile(options.OutputFile, []byte(response.Body), 0644)
		} else {
			err = response.SaveToFile(options.OutputFile)
		}

		if err != nil {
			return fmt.Errorf("failed to save output: %w", err)
		}

		fmt.Printf("Response saved to: %s\n", options.OutputFile)
	} else {
		response.Print()
	}

	return nil
}

func validateHTTPRequest(req *HttpRequest) error {
	if req.Method == "" {
		return fmt.Errorf("HTTP method is required")
	}

	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}

	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"HEAD": true, "OPTIONS": true, "PATCH": true, "TRACE": true,
	}

	if !validMethods[strings.ToUpper(req.Method)] {
		return fmt.Errorf("invalid HTTP method: %s", req.Method)
	}

	if !strings.Contains(req.URL, "://") && !strings.HasPrefix(req.URL, "/") {
		return fmt.Errorf("invalid URL format: %s", req.URL)
	}

	return nil
}
