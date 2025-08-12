// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"os"
	"rq/dock"
	"rq/request"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	parseArgs(args)
}

func printUsage() {
	fmt.Println("rq - A simple, powerful CLI for API testing")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rq dock init <name>         Create a new dock")
	fmt.Println("  rq dock list                List available docks")
	fmt.Println("  rq dock use <name>          Switch to a dock")
	fmt.Println("  rq new <name>               Create a new HTTP request")
	fmt.Println("  rq run <name>               Execute an HTTP request")
	fmt.Println("  rq run <name> --env <env>   Execute with specific environment")
	fmt.Println("  rq run <name> -o <file>     Save response to file")
}

func parseArgs(args []string) {
	switch args[0] {
	case "dock":
		if len(args) < 2 {
			fmt.Println("Error: dock command requires subcommand")
			os.Exit(1)
		}
		dock.Parse(args[1:])

	case "new":
		if len(args) < 2 {
			fmt.Println("Error: new command requires request name")
			os.Exit(1)
		}
		ctx := dock.GetContext()
		err := request.New(ctx, args[1], "http")
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created request: %s.http\n", args[1])

	case "run":
		if len(args) < 2 {
			fmt.Println("Error: run command requires request name")
			os.Exit(1)
		}

		ctx := dock.GetContext()
		requestName := args[1]

		// Parse flags
		options := request.ExecuteOptions{}
		i := 2
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
				i++
			default:
				fmt.Printf("Error: unknown flag %s\n", args[i])
				os.Exit(1)
			}
		}

		// Execute request
		var err error
		if options.Environment != "" || options.OutputFile != "" {
			err = request.EvaluateWithOptions(ctx, requestName, options)
		} else {
			err = request.Evaluate(ctx, requestName)
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Error: unknown command '%s'\n", args[0])
		printUsage()
		os.Exit(1)
	}
}
