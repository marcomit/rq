package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"rq/dock"

	"github.com/marcomit/args"
)

type DocComment struct {
	Type       string            // @doc, @param, @response, @example, etc.
	Content    string            // The main content
	Attributes map[string]string // Additional attributes
	LineNumber int               // Line number in file
}

type RequestDoc struct {
	Name         string        // Request name (filename without extension)
	FilePath     string        // Full file path
	RelativePath string        // Path relative to dock
	Method       string        // HTTP method
	URL          string        // Request URL pattern
	Description  string        // Main description
	Parameters   []ParamDoc    // Request parameters
	Headers      []HeaderDoc   // Request headers
	Responses    []ResponseDoc // Response documentation
	Examples     []ExampleDoc  // Usage examples
	Tags         []string      // Categories/tags
	Since        string        // Version since when available
	Deprecated   bool          // Whether deprecated
	Comments     []DocComment  // All parsed comments
	RequestBody  string        // Example request body
}

type ParamDoc struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Default     string `json:"default"`
}

type HeaderDoc struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

type ResponseDoc struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	ContentType string `json:"content_type"`
	Example     string `json:"example"`
	Schema      string `json:"schema"`
}

type ExampleDoc struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Output      string `json:"output"`
}

type DockDocs struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Version     string                  `json:"version"`
	BaseURL     string                  `json:"base_url"`
	Requests    []RequestDoc            `json:"requests"`
	Groups      map[string][]RequestDoc `json:"groups"`
	GeneratedAt time.Time               `json:"generated_at"`
	DockPath    string                  `json:"dock_path"`
}

func Setup(app *args.Parser) {
	docs := app.Command("docs", "Manage the documentation of the dock")

	docs.
		Command("generate", "Generate the documentation").
		Option("output", "o", "Output path of the documentation")

	docs.
		Command("serve", "Serve the documentation as webapp").
		Option("port", "p", "Server port")

	docs.
		Command("export", "Export documentation").
		Option("output", "o", "Output path of the documentation").
		Option("format", "format", "Format type of the documentation")

}

func Parse(args []string) {
	if len(args) == 0 {
		generateDocs("")
		return
	}

	switch args[0] {
	case "generate", "gen":
		output := ""
		if len(args) > 1 {
			output = args[1]
		}
		generateDocs(output)

	case "serve":
		port := "8080"
		if len(args) > 1 {
			port = args[1]
		}
		serveDocs(port)

	case "export":
		format := "html"
		output := ""
		if len(args) > 1 {
			format = args[1]
		}
		if len(args) > 2 {
			output = args[2]
		}
		exportDocs(format, output)

	case "--help", "-h":
		printDocsHelp()

	default:
		fmt.Printf("Error: unknown docs subcommand '%s'\n", args[0])
		printDocsHelp()
		os.Exit(1)
	}
}

func printDocsHelp() {
	fmt.Println("Usage: rq docs <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  generate [output]     Generate documentation (default: stdout)")
	fmt.Println("  serve [port]          Serve documentation on HTTP server (default: 8080)")
	fmt.Println("  export <format> [output]  Export docs in different formats")
	fmt.Println()
	fmt.Println("Export formats:")
	fmt.Println("  html                  HTML documentation")
	fmt.Println("  markdown, md          Markdown documentation")
	fmt.Println("  json                  JSON documentation")
	fmt.Println("  openapi               OpenAPI 3.0 specification")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  rq docs generate")
	fmt.Println("  rq docs serve 3000")
	fmt.Println("  rq docs export html docs.html")
	fmt.Println("  rq docs export openapi api-spec.yaml")
}

func generateDocs(output string) {
	ctx := dock.GetContext()

	dockDocs, err := extractDockDocs(ctx)
	if err != nil {
		fmt.Printf("Error extracting documentation: %v\n", err)
		os.Exit(1)
	}

	if output == "" {
		printDocsToStdout(dockDocs)
	} else {
		err := saveDocs(dockDocs, output)
		if err != nil {
			fmt.Printf("Error saving documentation: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Documentation generated: %s\n", output)
	}
}

func extractDockDocs(ctx *dock.RqContext) (*DockDocs, error) {
	dockDocs := &DockDocs{
		DockPath:    ctx.Dock,
		GeneratedAt: time.Now(),
		Groups:      make(map[string][]RequestDoc),
	}

	dockFile := filepath.Join(ctx.Dock, ".dock")
	if content, err := os.ReadFile(dockFile); err == nil {
		dockDocs.Name = strings.TrimSpace(string(content))
	} else {
		dockDocs.Name = filepath.Base(ctx.Dock)
	}

	if readme := filepath.Join(ctx.Dock, "README.md"); fileExists(readme) {
		if content, err := os.ReadFile(readme); err == nil {
			dockDocs.Description = extractFirstParagraph(string(content))
		}
	}

	if config, err := ctx.GetConfig(""); err == nil {
		if baseURL, ok := config["BASE_URL"]; ok {
			dockDocs.BaseURL = baseURL
		}
		if version, ok := config["API_VERSION"]; ok {
			dockDocs.Version = version
		}
	}

	err := filepath.Walk(ctx.Dock, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".http") ||
			strings.HasSuffix(path, ".ws") || strings.HasSuffix(path, ".grpc")) {

			reqDoc, err := extractRequestDoc(path, ctx.Dock)
			if err != nil {
				fmt.Printf("Warning: failed to parse %s: %v\n", path, err)
				return nil
			}

			dockDocs.Requests = append(dockDocs.Requests, reqDoc)

			dir := filepath.Dir(reqDoc.RelativePath)
			if dir == "." {
				dir = "Root"
			}
			dockDocs.Groups[dir] = append(dockDocs.Groups[dir], reqDoc)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk dock directory: %w", err)
	}

	sort.Slice(dockDocs.Requests, func(i, j int) bool {
		return dockDocs.Requests[i].Name < dockDocs.Requests[j].Name
	})

	return dockDocs, nil
}

func extractRequestDoc(filePath, dockPath string) (RequestDoc, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return RequestDoc{}, fmt.Errorf("failed to read file: %w", err)
	}

	relPath, _ := filepath.Rel(dockPath, filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	reqDoc := RequestDoc{
		Name:         name,
		FilePath:     filePath,
		RelativePath: relPath,
		Parameters:   []ParamDoc{},
		Headers:      []HeaderDoc{},
		Responses:    []ResponseDoc{},
		Examples:     []ExampleDoc{},
		Tags:         []string{},
		Comments:     []DocComment{},
	}

	lines := strings.Split(string(content), "\n")

	inDocBlock := false
	currentDocBlock := []string{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "///") || strings.HasPrefix(trimmed, "##") {
			if !inDocBlock {
				inDocBlock = true
				currentDocBlock = []string{}
			}

			docLine := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "///"), "##"))
			currentDocBlock = append(currentDocBlock, docLine)
		} else if inDocBlock && (trimmed == "" || strings.HasPrefix(trimmed, "#")) {
			if strings.HasPrefix(trimmed, "#") {
				docLine := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
				currentDocBlock = append(currentDocBlock, docLine)
			} else {
				currentDocBlock = append(currentDocBlock, "")
			}
		} else {
			if inDocBlock {
				processDocBlock(currentDocBlock, &reqDoc, i)
				inDocBlock = false
				currentDocBlock = []string{}
			}

			if reqDoc.Method == "" && reqDoc.URL == "" {
				if method, url := parseHTTPRequestLine(trimmed); method != "" {
					reqDoc.Method = method
					reqDoc.URL = url
				}
			}

			if trimmed != "" && !strings.Contains(trimmed, ":") &&
				reqDoc.Method != "" && strings.HasPrefix(trimmed, "{") {
				reqDoc.RequestBody = captureJSONBody(lines[i:])
				break
			}
		}
	}

	if inDocBlock {
		processDocBlock(currentDocBlock, &reqDoc, len(lines))
	}

	return reqDoc, nil
}

func processDocBlock(lines []string, reqDoc *RequestDoc, lineNum int) {
	content := strings.Join(lines, "\n")
	content = strings.TrimSpace(content)

	if content == "" {
		return
	}

	docLines := strings.Split(content, "\n")

	for _, line := range docLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		comment := DocComment{
			Content:    line,
			LineNumber: lineNum,
			Attributes: make(map[string]string),
		}

		if strings.HasPrefix(line, "@") {
			parseDocTag(line, reqDoc, &comment)
		} else if reqDoc.Description == "" {
			reqDoc.Description = line
		}

		reqDoc.Comments = append(reqDoc.Comments, comment)
	}
}

func parseDocTag(line string, reqDoc *RequestDoc, comment *DocComment) {
	tagRegex := regexp.MustCompile(`^@(\w+)(?:\(([^)]+)\))?\s*(.*)$`)
	matches := tagRegex.FindStringSubmatch(line)

	if len(matches) < 2 {
		return
	}

	tag := matches[1]
	attributes := matches[2]
	content := strings.TrimSpace(matches[3])

	comment.Type = tag
	comment.Content = content

	if attributes != "" {
		parseAttributes(attributes, comment.Attributes)
	}

	switch tag {
	case "doc", "description":
		if content != "" {
			reqDoc.Description = content
		}

	case "param", "parameter":
		param := ParamDoc{
			Name:        comment.Attributes["name"],
			Type:        comment.Attributes["type"],
			Description: content,
			Example:     comment.Attributes["example"],
			Default:     comment.Attributes["default"],
			Required:    comment.Attributes["required"] == "true",
		}
		if param.Name != "" {
			reqDoc.Parameters = append(reqDoc.Parameters, param)
		}

	case "header":
		header := HeaderDoc{
			Name:        comment.Attributes["name"],
			Description: content,
			Example:     comment.Attributes["example"],
			Required:    comment.Attributes["required"] == "true",
		}
		if header.Name != "" {
			reqDoc.Headers = append(reqDoc.Headers, header)
		}

	case "response":
		response := ResponseDoc{
			Status:      comment.Attributes["status"],
			Description: content,
			ContentType: comment.Attributes["type"],
			Example:     comment.Attributes["example"],
			Schema:      comment.Attributes["schema"],
		}
		if response.Status != "" {
			reqDoc.Responses = append(reqDoc.Responses, response)
		}

	case "example":
		example := ExampleDoc{
			Title:       comment.Attributes["title"],
			Description: content,
			Code:        comment.Attributes["code"],
			Output:      comment.Attributes["output"],
		}
		reqDoc.Examples = append(reqDoc.Examples, example)

	case "tag", "tags":
		if content != "" {
			tags := strings.Split(content, ",")
			for _, tag := range tags {
				reqDoc.Tags = append(reqDoc.Tags, strings.TrimSpace(tag))
			}
		}

	case "since":
		reqDoc.Since = content

	case "deprecated":
		reqDoc.Deprecated = true
	}
}

func parseAttributes(attrStr string, attrs map[string]string) {
	pairs := strings.Split(attrStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.Trim(strings.TrimSpace(kv[1]), `"'`)
			attrs[key] = value
		}
	}
}

func parseHTTPRequestLine(line string) (method, url string) {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		method = strings.ToUpper(parts[0])
		url = parts[1]

		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true,
			"HEAD": true, "OPTIONS": true, "PATCH": true, "TRACE": true,
		}

		if validMethods[method] {
			return method, url
		}
	}
	return "", ""
}

func captureJSONBody(lines []string) string {
	var body strings.Builder
	braceCount := 0
	started := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" && !started {
			continue
		}

		if strings.HasPrefix(trimmed, "{") || started {
			started = true
			body.WriteString(line + "\n")

			for _, char := range trimmed {
				if char == '{' {
					braceCount++
				} else if char == '}' {
					braceCount--
				}
			}

			if braceCount == 0 && started {
				break
			}
		}
	}

	return strings.TrimSpace(body.String())
}

func printDocsToStdout(dockDocs *DockDocs) {
	fmt.Printf("# %s API Documentation\n\n", dockDocs.Name)

	if dockDocs.Description != "" {
		fmt.Printf("%s\n\n", dockDocs.Description)
	}

	if dockDocs.BaseURL != "" {
		fmt.Printf("**Base URL:** `%s`\n\n", dockDocs.BaseURL)
	}

	if dockDocs.Version != "" {
		fmt.Printf("**Version:** %s\n\n", dockDocs.Version)
	}

	fmt.Printf("**Generated:** %s\n\n", dockDocs.GeneratedAt.Format("2006-01-02 15:04:05"))

	for groupName, requests := range dockDocs.Groups {
		fmt.Printf("## %s\n\n", groupName)

		for _, req := range requests {
			printRequestDoc(req)
		}
	}
}

func printRequestDoc(req RequestDoc) {
	fmt.Printf("### %s\n\n", req.Name)

	if req.Method != "" && req.URL != "" {
		fmt.Printf("**`%s %s`**\n\n", req.Method, req.URL)
	}

	if req.Description != "" {
		fmt.Printf("%s\n\n", req.Description)
	}

	if req.Deprecated {
		fmt.Printf("⚠️ **DEPRECATED**\n\n")
	}

	if len(req.Tags) > 0 {
		fmt.Printf("**Tags:** %s\n\n", strings.Join(req.Tags, ", "))
	}

	if len(req.Parameters) > 0 {
		fmt.Println("**Parameters:**")
		fmt.Println("| Name | Type | Required | Description | Example |")
		fmt.Println("|------|------|----------|-------------|---------|")
		for _, param := range req.Parameters {
			required := "No"
			if param.Required {
				required = "Yes"
			}
			fmt.Printf("| %s | %s | %s | %s | %s |\n",
				param.Name, param.Type, required, param.Description, param.Example)
		}
		fmt.Println()
	}

	if len(req.Responses) > 0 {
		fmt.Println("**Responses:**")
		for _, resp := range req.Responses {
			fmt.Printf("- **%s**: %s\n", resp.Status, resp.Description)
			if resp.Example != "" {
				fmt.Printf("  ```json\n  %s\n  ```\n", resp.Example)
			}
		}
		fmt.Println()
	}

	if req.RequestBody != "" {
		fmt.Println("**Request Body:**")
		fmt.Printf("```json\n%s\n```\n\n", req.RequestBody)
	}

	fmt.Println("---\n")
}

func saveDocs(dockDocs *DockDocs, output string) error {
	content := generateMarkdownDocs(dockDocs)
	return os.WriteFile(output, []byte(content), 0644)
}

func generateMarkdownDocs(dockDocs *DockDocs) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# %s API Documentation\n\n", dockDocs.Name))

	if dockDocs.Description != "" {
		md.WriteString(fmt.Sprintf("%s\n\n", dockDocs.Description))
	}

	if dockDocs.BaseURL != "" {
		md.WriteString(fmt.Sprintf("**Base URL:** `%s`\n\n", dockDocs.BaseURL))
	}

	if dockDocs.Version != "" {
		md.WriteString(fmt.Sprintf("**Version:** %s\n\n", dockDocs.Version))
	}

	md.WriteString(fmt.Sprintf("**Generated:** %s\n\n", dockDocs.GeneratedAt.Format("2006-01-02 15:04:05")))

	md.WriteString("## Table of Contents\n\n")
	for groupName, requests := range dockDocs.Groups {
		md.WriteString(fmt.Sprintf("- [%s](#%s)\n", groupName,
			strings.ToLower(strings.ReplaceAll(groupName, " ", "-"))))
		for _, req := range requests {
			md.WriteString(fmt.Sprintf("  - [%s](#%s)\n", req.Name,
				strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))))
		}
	}
	md.WriteString("\n")

	for groupName, requests := range dockDocs.Groups {
		md.WriteString(fmt.Sprintf("## %s\n\n", groupName))

		for _, req := range requests {
			md.WriteString(generateRequestMarkdown(req))
		}
	}

	return md.String()
}

func generateRequestMarkdown(req RequestDoc) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("### %s\n\n", req.Name))

	if req.Method != "" && req.URL != "" {
		md.WriteString(fmt.Sprintf("**`%s %s`**\n\n", req.Method, req.URL))
	}

	if req.Description != "" {
		md.WriteString(fmt.Sprintf("%s\n\n", req.Description))
	}

	if req.Deprecated {
		md.WriteString("⚠️ **DEPRECATED**\n\n")
	}

	if len(req.Tags) > 0 {
		md.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(req.Tags, ", ")))
	}

	if len(req.Parameters) > 0 {
		md.WriteString("**Parameters:**\n\n")
		md.WriteString("| Name | Type | Required | Description | Example |\n")
		md.WriteString("|------|------|----------|-------------|----------|\n")
		for _, param := range req.Parameters {
			required := "No"
			if param.Required {
				required = "Yes"
			}
			md.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				param.Name, param.Type, required, param.Description, param.Example))
		}
		md.WriteString("\n")
	}

	if len(req.Headers) > 0 {
		md.WriteString("**Headers:**\n\n")
		md.WriteString("| Name | Required | Description | Example |\n")
		md.WriteString("|------|----------|-------------|----------|\n")
		for _, header := range req.Headers {
			required := "No"
			if header.Required {
				required = "Yes"
			}
			md.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				header.Name, required, header.Description, header.Example))
		}
		md.WriteString("\n")
	}

	if len(req.Responses) > 0 {
		md.WriteString("**Responses:**\n\n")
		for _, resp := range req.Responses {
			md.WriteString(fmt.Sprintf("- **%s**: %s\n", resp.Status, resp.Description))
			if resp.Example != "" {
				md.WriteString(fmt.Sprintf("  ```json\n  %s\n  ```\n", resp.Example))
			}
		}
		md.WriteString("\n")
	}

	if req.RequestBody != "" {
		md.WriteString("**Request Body:**\n\n")
		md.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", req.RequestBody))
	}

	if len(req.Examples) > 0 {
		md.WriteString("**Examples:**\n\n")
		for _, example := range req.Examples {
			if example.Title != "" {
				md.WriteString(fmt.Sprintf("#### %s\n\n", example.Title))
			}
			if example.Description != "" {
				md.WriteString(fmt.Sprintf("%s\n\n", example.Description))
			}
			if example.Code != "" {
				md.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", example.Code))
			}
			if example.Output != "" {
				md.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", example.Output))
			}
		}
	}

	md.WriteString("---\n\n")

	return md.String()
}

func serveDocs(port string) {
	fmt.Printf("Documentation server not yet implemented\n")
	fmt.Printf("Will serve on http://localhost:%s\n", port)
}

func exportDocs(format, output string) {
	fmt.Printf("Export to %s format not yet implemented\n", format)
	if output != "" {
		fmt.Printf("Would save to: %s\n", output)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func extractFirstParagraph(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}
