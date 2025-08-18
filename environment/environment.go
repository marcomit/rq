package environment

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"rq/dock"
	"strings"

	"github.com/marcomit/args"
)

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

func List() error {
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
}

func Show(path string) error {
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
}

func Setup(app *args.Parser) {
	env := app.Command("env", "Environment manager")

	env.Command("list", "Shows all environments in the current dock").
		Action(func(r *args.Result) error {
			return List()
		})

	env.Command("show", "Shows the current configuration").
		Positional("path").
		Action(func(r *args.Result) error {
			if len(r.Positionals) == 0 {
				return errors.New("Missing path argument")
			}
			return Show(r.Positionals[0])
		})
}
