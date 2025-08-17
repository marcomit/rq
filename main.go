// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"rq/dock"
	"rq/request"
	"strconv"
	"strings"
	"time"

	argparser "github.com/marcomit/args"
)

func main() {
	rq := argparser.New("rq")

	setupRequest(rq)
	setupDock(rq)
	setupEnv(rq)
	setupDocs(rq)

	err := rq.Run(os.Args[1:])
	if err != nil {
		fmt.Println(err)
	}
}

func setupDock(app *argparser.Parser) {
	dockCmd := app.Command("dock", "Dock command")

	dockCmd.Command("init", "Initialize an rq dock").Positional("name").
		Action(func(r *argparser.Result) error {
			if len(r.Positionals) == 0 {
				return errors.New("Expected one positional argument")
			}
			dock.CreateDock(r.Positionals[0])
			return nil
		})

	dockCmd.Command("use", "Change the active dock").Positional("name").
		Action(func(r *argparser.Result) error {
			if len(r.Positionals) == 0 {
				return errors.New("Expected one positional argument")
			}
			dock.SetCurrentDock(r.Positionals[0])
			return nil
		})

	dockCmd.Command("list", "Lists all rq docks").
		Action(func(r *argparser.Result) error {
			dock.List()
			return nil
		})

	dockCmd.Command("status", "Check the status of the dock").
		Action(func(r *argparser.Result) error {
			dock.ShowStatus()
			return nil
		})
}

func setupEnv(app *argparser.Parser) {
	env := app.Command("env", "Manage the environment")

	env.Command("list", "Shows all environments in the current dock").
		Action(func(r *argparser.Result) error {
			ctx := dock.GetContext()

			fmt.Printf("Environment files in dock: %s\n", ctx.Dock)
			fmt.Println()

			envFiles := findEnvFiles(ctx.Dock)

			if len(envFiles) == 0 {
				return errors.New("No environment files found")
			}

			for _, envFile := range envFiles {
				relPath, _ := filepath.Rel(ctx.Dock, envFile)
				fmt.Printf("  %s\n", relPath)
			}
			return nil
		})

	env.Command("show", "Shows the current configuration").
		Positional("path").
		Action(func(r *argparser.Result) error {
			if len(r.Positionals) == 0 {
				return errors.New("Missing path argument")
			}
			path := r.Positionals[0]
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
				return errors.New("No configuration variables found")
			}

			for key, value := range config {
				fmt.Printf("  %s=%s\n", key, value)
			}
			return nil
		})
}

func setupRequest(app *argparser.Parser) {
	app.
		Command("run", "Runs the specified request").
		Positional("name").
		Option("env", "e", "Environment").
		Option("output", "o", "Choose the file to write the response").
		Option("timeout", "t", "Set the timeout to abort the request").
		Flag("output-body", "ob", "If flagged it saves only the body (avoid saving headers)").
		Action(func(r *argparser.Result) error {
			if len(r.Positionals) == 0 {
				return errors.New("Missing name of the request to run")
			}
			name := r.Positionals[0]

			options := request.ExecuteOptions{
				Timeout: 30 * time.Second,
			}

			if env, ok := r.Options["env"]; ok {
				options.Environment = env
			}

			if output, ok := r.Options["output"]; ok {
				options.OutputFile = output
			}
			if r.Flag("output-body") {
				options.OutputBodyOnly = true
			}

			if timeout, ok := r.Options["timeout"]; ok {
				val, err := strconv.Atoi(timeout)
				if err != nil {
					return errors.New("Timeout must be a number")
				}
				options.Timeout = (time.Duration(val) * time.Second)
			}

			ctx := dock.GetContext()

			var err error
			if options.Environment != "" || options.OutputFile != "" || options.Timeout != 30*time.Second {
				err = request.EvaluateWithOptions(ctx, name, options)
			} else {
				err = request.Evaluate(ctx, name)
			}
			return err
		})

	app.Command("new", "Create a new request").
		Positional("name").
		Option("protocol", "p", "Set the protocol for the request", "http", "tcp").
		Action(func(r *argparser.Result) error {
			if len(r.Positionals) == 0 {
				return errors.New("Missing name of the request")
			}
			name := r.Positionals[0]
			protocol := "http"
			if _, ok := r.Options["protocol"]; ok {
				protocol = r.Options["protocol"]
			}

			ctx := dock.GetContext()
			err := request.New(ctx, name, protocol)
			if err != nil {
				fmt.Printf("Error creating request: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Created request: %s.%s\n", name, protocol)
			fmt.Printf("Edit the file to customize your request\n")
			return nil
		})
}

func setupDocs(app *argparser.Parser) {
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
