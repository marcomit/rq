// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"rq/dock"
	"rq/docs"
	"rq/request"
	"strconv"
	"strings"
	"time"
)

const VERSION = "1.0.0"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	parseArgs(args)
}

func printUsage() {
	fmt.Printf("rq v%s - A simple, powerful CLI for API testing\n", VERSION)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rq dock init <n>                  Create a new dock")
	fmt.Println("  rq dock list                         List available docks")
	fmt.Println("  rq dock use <n>                   Switch to a dock")
	fmt.Println("  rq dock status                       Show current dock status")
	fmt.Println()
	fmt.Println("  rq new <n>                        Create a new HTTP request")
	fmt.Println("  rq new <path/name>                   Create request in subdock")
	fmt.Println()
	fmt.Println("  rq run <n>                        Execute an HTTP request")
	fmt.Println("  rq run <n> --env <env>            Execute with specific environment")
	fmt.Println("  rq run <n> -o <file>              Save response to file")
	fmt.Println("  rq run <n> --output-body <file>   Save only response body")
	fmt.Println("  rq run <n> --timeout <seconds>    Set request timeout")
	fmt.Println()
	fmt.Println("  rq docs [generate]                   Generate API documentation")
	fmt.Println("  rq docs serve [port]                 Serve docs on HTTP server")
	fmt.Println("  rq docs export <format> [file]      Export docs in different formats")
	fmt.Println()
	fmt.Println("  rq env list                          Show available environments")
	fmt.Println("  rq env show [path]                   Show effective configuration")
	fmt.Println()
	fmt.Println("  rq version                           Show version information")
	fmt.Println("  rq help                              Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  rq dock init my-api                  Create 'my-api' dock")
	fmt.Println("  rq new login                         Create login.http request")
	fmt.Println("  rq new auth/oauth                    Create oauth.http in auth/ subdock")
	fmt.Println("  rq run login --env prod              Run login request with production env")
	fmt.Println("  rq run users -o response.json        Save response to file")
	fmt.Println("  rq docs generate docs.md             Generate documentation")
	fmt.Println("  rq docs serve 3000                   Serve docs on port 3000")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/marcomit/rq")
}

func printVersion() {
	fmt.Printf("rq version %s\n", VERSION)
	fmt.Println("A simple, powerful CLI for API testing")
	fmt.Println("https://github.com/marcomit/rq")
}

func parseArgs(args []string) {
	command := args[0]

	switch command {
	case "version", "--version", "-v":
		printVersion()

	case "help", "--help", "-h":
		printUsage()

	case "dock":
		if len(args) < 2 {
			fmt.Println("Error: dock command requires subcommand")
			fmt.Println("Available subcommands: init, list, use, status")
			os.Exit(1)
		}
		dock.Parse(args[1:])

	case "new":
		if len(args) < 2 {
			fmt.Println("Error: new command requires request name")
			fmt.Println("Usage: rq new <n>")
			os.Exit(1)
		}
		handleNewCommand(args[1:])

	case "run":
		if len(args) < 2 {
			fmt.Println("Error: run command requires request name")
			fmt.Println("Usage: rq run <n> [options]")
			os.Exit(1)
		}
		handleRunCommand(args[1:])

	case "docs":
		docs.Parse(args[1:])

	case "env":
		if len(args) < 2 {
			fmt.Println("Error: env command requires subcommand")
			fmt.Println("Available subcommands: list, show")
			os.Exit(1)
		}
		handleEnvCommand(args[1:])

	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		fmt.Println("Run 'rq help' to see available commands")
		os.Exit(1)
	}
}

func handleNewCommand(args []string) {
	requestName := args[0]
	protocol := "http"

	i := 1
	for i < len(args) {
		switch args[i] {
		case "--type":
			if i+1 >= len(args) {
				fmt.Println("Error: --type requires protocol type")
				os.Exit(1)
			}
			protocol = args[i+1]
			i += 2
		case "--help", "-h":
			fmt.Println("Usage: rq new <n> [options]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --type <protocol>    Protocol type (default: http)")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  rq new login")
			fmt.Println("  rq new auth/signin")
			fmt.Println("  rq new chat --type ws")
			return
		default:
			fmt.Printf("Error: unknown flag '%s'\n", args[i])
			fmt.Println("Run 'rq new --help' for usage information")
			os.Exit(1)
		}
	}

	ctx := dock.GetContext()
	err := request.New(ctx, requestName, protocol)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created request: %s.%s\n", requestName, protocol)
	fmt.Printf("Edit the file to customize your request\n")

	if protocol == "http" {
		fmt.Println()
		fmt.Println("💡 Tip: Add documentation using special comments:")
		fmt.Println("   /// @doc This endpoint retrieves user information")
		fmt.Println("   /// @param(name=id, type=integer, required=true) User ID")
		fmt.Println("   /// @response(status=200) User data returned successfully")
		fmt.Println()
		fmt.Println("   Then generate docs with: rq docs generate")
	}
}

func handleRunCommand(args []string) {
	requestName := args[0]

	options := request.ExecuteOptions{
		Timeout: 30 * time.Second,
	}

	i := 1
	for i < len(args) {
		switch args[i] {
		case "--env":
			if i+1 >= len(args) {
				fmt.Println("Error: --env requires environment name")
				os.Exit(1)
			}
			options.Environment = args[i+1]
			i += 2

		case "-o", "--output":
			if i+1 >= len(args) {
				fmt.Println("Error: -o requires output file")
				os.Exit(1)
			}
			options.OutputFile = args[i+1]
			i += 2

		case "--output-body":
			options.OutputBodyOnly = true
			if i+1 < len(args) && !isFlag(args[i+1]) {
				options.OutputFile = args[i+1]
				i += 2
			} else {
				i++
			}

		case "--timeout":
			if i+1 >= len(args) {
				fmt.Println("Error: --timeout requires timeout in seconds")
				os.Exit(1)
			}
			timeout, err := strconv.Atoi(args[i+1])
			if err != nil || timeout <= 0 {
				fmt.Printf("Error: invalid timeout '%s'. Must be a positive integer\n", args[i+1])
				os.Exit(1)
			}
			options.Timeout = time.Duration(timeout) * time.Second
			i += 2

		case "--help", "-h":
			fmt.Println("Usage: rq run <n> [options]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --env <env>              Use specific environment")
			fmt.Println("  -o, --output <file>      Save response to file")
			fmt.Println("  --output-body [file]     Save only response body")
			fmt.Println("  --timeout <seconds>      Request timeout (default: 30)")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  rq run login")
			fmt.Println("  rq run users --env prod")
			fmt.Println("  rq run api-call -o response.json")
			fmt.Println("  rq run upload --timeout 60")
			return

		default:
			fmt.Printf("Error: unknown flag '%s'\n", args[i])
			fmt.Println("Run 'rq run --help' for usage information")
			os.Exit(1)
		}
	}

	ctx := dock.GetContext()

	var err error
	if options.Environment != "" || options.OutputFile != "" || options.Timeout != 30*time.Second {
		err = request.EvaluateWithOptions(ctx, requestName, options)
	} else {
		err = request.Evaluate(ctx, requestName)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleEnvCommand(args []string) {
	subcommand := args[0]

	switch subcommand {
	case "list":
		handleEnvList()
	case "show":
		path := ""
		if len(args) > 1 {
			path = args[1]
		}
		handleEnvShow(path)
	case "--help", "-h":
		fmt.Println("Usage: rq env <subcommand> [options]")
		fmt.Println()
		fmt.Println("Subcommands:")
		fmt.Println("  list           Show available environments")
		fmt.Println("  show [path]    Show effective configuration")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  rq env list")
		fmt.Println("  rq env show")
		fmt.Println("  rq env show auth")
	default:
		fmt.Printf("Error: unknown env subcommand '%s'\n", subcommand)
		fmt.Println("Available subcommands: list, show")
		os.Exit(1)
	}
}

func handleEnvList() {
	ctx := dock.GetContext()

	fmt.Printf("Environment files in dock: %s\n", ctx.Dock)
	fmt.Println()

	envFiles := findEnvFiles(ctx.Dock)

	if len(envFiles) == 0 {
		fmt.Println("No environment files found")
		return
	}

	for _, envFile := range envFiles {
		relPath, _ := filepath.Rel(ctx.Dock, envFile)
		fmt.Printf("  %s\n", relPath)
	}
}

func handleEnvShow(path string) {
	ctx := dock.GetContext()

	config, err := ctx.GetConfig(path)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if path == "" {
		fmt.Printf("Configuration for dock root:\n")
	} else {
		fmt.Printf("Configuration for path '%s':\n", path)
	}
	fmt.Println()

	if len(config) == 0 {
		fmt.Println("No configuration variables found")
		return
	}

	for key, value := range config {
		fmt.Printf("  %s=%s\n", key, value)
	}
}

func isFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

func findEnvFiles(root string) []string {
	var envFiles []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && (info.Name() == ".env" || strings.HasPrefix(info.Name(), ".env.")) {
			envFiles = append(envFiles, path)
		}

		return nil
	})

	return envFiles
}
