// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package dock

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SetCurrentDock(name string) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		fmt.Printf("Error: dock '%s' does not exist\n", name)
		os.Exit(1)
	}

	dockFile := filepath.Join(name, ".dock")
	if _, err := os.Stat(dockFile); os.IsNotExist(err) {
		fmt.Printf("Error: '%s' is not a valid dock (missing .dock file)\n", name)
		os.Exit(1)
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error: failed to get config directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(dir, "rq")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error: failed to create config directory: %v\n", err)
		os.Exit(1)
	}

	configFile := filepath.Join(configDir, "current_dock")
	absPath, err := filepath.Abs(name)
	if err != nil {
		fmt.Printf("Error: failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configFile, []byte(absPath), 0644); err != nil {
		fmt.Printf("Error: failed to set current dock: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Switched to dock: %s\n", name)
}

func CreateDock(name string) {
	fmt.Printf("Creating dock '%s'...\n", name)

	if _, err := os.Stat(name); err == nil {
		fmt.Printf("Error: directory '%s' already exists\n", name)
		os.Exit(1)
	}

	if err := os.Mkdir(name, 0755); err != nil {
		fmt.Printf("Error: failed to create dock directory: %v\n", err)
		os.Exit(1)
	}

	dockFile := filepath.Join(name, ".dock")
	dock, err := os.Create(dockFile)
	if err != nil {
		fmt.Printf("Error: failed to create .dock file: %v\n", err)
		os.RemoveAll(name)
		os.Exit(1)
	}
	defer dock.Close()

	if _, err := dock.WriteString(name); err != nil {
		fmt.Printf("Error: failed to write dock name: %v\n", err)
		os.RemoveAll(name)
		os.Exit(1)
	}

	envFile := filepath.Join(name, ".env")
	env, err := os.Create(envFile)
	if err != nil {
		fmt.Printf("Error: failed to create environment file: %v\n", err)
		os.RemoveAll(name)
		os.Exit(1)
	}
	defer env.Close()

	defaultEnv := `# RQ Environment Configuration
# Base URL for your API
BASE_URL=https://api.example.com

# HTTP Version (default: HTTP/1.1)
HTTP_VERSION=HTTP/1.1

# Add your custom variables below
# API_KEY=your_api_key_here
# JWT_TOKEN=your_jwt_token_here
`

	if _, err := env.WriteString(defaultEnv); err != nil {
		fmt.Printf("Error: failed to write default environment: %v\n", err)
		os.RemoveAll(name)
		os.Exit(1)
	}

	fmt.Printf("Successfully created dock '%s'\n", name)
	fmt.Println("Edit the .env file to configure your environment variables")
}

func List() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	docks := findDocks(wd)

	if len(docks) == 0 {
		fmt.Println("No docks found in current directory and subdirectories")
		return
	}

	fmt.Println("Available docks:")
	for _, dock := range docks {
		dockFile := filepath.Join(dock, ".dock")
		content, err := os.ReadFile(dockFile)
		if err != nil {
			fmt.Printf("  %s (error reading name)\n", dock)
			continue
		}

		name := strings.TrimSpace(string(content))
		if name == "" {
			name = filepath.Base(dock)
		}

		fmt.Printf("  %s (%s)\n", name, dock)
	}
}

func findDocks(root string) []string {
	var docks []string

	entries, err := os.ReadDir(root)
	if err != nil {
		return docks
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(root, entry.Name())
		dockFile := filepath.Join(dirPath, ".dock")

		if _, err := os.Stat(dockFile); err == nil {
			docks = append(docks, dirPath)
		}

		subdocks := findDocks(dirPath)
		docks = append(docks, subdocks...)
	}

	return docks
}

func ShowStatus() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	ctx := &RqContext{Path: wd}

	if !ctx.IsValidDock() {
		fmt.Printf("Current directory is not a valid dock: %s\n", wd)
		fmt.Println("Run 'rq dock init <name>' to create a new dock")
		return
	}

	root, err := ctx.GetDockRoot()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	dockFile := filepath.Join(root, ".dock")
	content, err := os.ReadFile(dockFile)
	if err != nil {
		fmt.Printf("Error reading dock file: %v\n", err)
		os.Exit(1)
	}

	name := strings.TrimSpace(string(content))
	if name == "" {
		name = filepath.Base(root)
	}

	fmt.Printf("Current dock: %s\n", name)
	fmt.Printf("Dock path: %s\n", root)
	fmt.Printf("Working directory: %s\n", wd)

	requests := findRequests(wd)
	if len(requests) > 0 {
		fmt.Println("Available requests:")
		for _, req := range requests {
			relPath, _ := filepath.Rel(root, req)
			reqName := strings.TrimSuffix(filepath.Base(req), ".http")
			fmt.Printf("  %s (%s)\n", reqName, relPath)
		}
	} else {
		fmt.Println("No requests found")
		fmt.Println("Run 'rq new <name>' to create a new request")
	}
}

func findRequests(root string) []string {
	var requests []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(path, ".http") {
			requests = append(requests, path)
		}

		return nil
	})

	return requests
}
