// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package dock

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

type RqContext struct {
	Path string
	Dock string
}

func exists(path string) bool {
	_, err := os.Stat(filepath.Clean(path))
	return err == nil
}

func validatePath(path string, predicate func(string) bool) []string {
	res := []string{}

	for {
		if predicate(path) {
			res = append(res, path)
		}

		parent := filepath.Dir(path)

		if parent == path {
			break
		}

		path = parent
	}

	return res
}

func (ctx *RqContext) IsValidDock() bool {
	res := validatePath(ctx.Path, func(curr string) bool {
		path := filepath.Join(curr, ".dock")
		return exists(path)
	})

	return len(res) > 0
}

func (ctx *RqContext) GetDockRoot() (string, error) {
	res := validatePath(ctx.Path, func(curr string) bool {
		return exists(filepath.Join(curr, ".dock"))
	})

	if len(res) == 0 {
		return "", fmt.Errorf("no valid dock path found")
	}

	return res[0], nil
}

func loadConfig(path string) (map[string]string, error) {
	res := make(map[string]string)

	file, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return res, nil
		}
		return res, err
	}

	lines := strings.Split(string(file), "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return res, fmt.Errorf("invalid format at line %d: missing '=' character", lineNum+1)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return res, fmt.Errorf("empty key at line %d", lineNum+1)
		}

		res[key] = value
	}

	return res, nil
}

func (ctx *RqContext) GetConfig(relpath string) (map[string]string, error) {
	configs := make(map[string]string)

	rootConfigPath := filepath.Join(ctx.Dock, ".env")
	rootConfig, err := loadConfig(rootConfigPath)
	if err != nil {
		return configs, fmt.Errorf("failed to load root config: %w", err)
	}
	maps.Copy(configs, rootConfig)

	if relpath == "" {
		return configs, nil
	}

	currentPath := ctx.Dock
	pathSegments := strings.Split(strings.Trim(relpath, string(os.PathSeparator)), string(os.PathSeparator))

	for _, segment := range pathSegments {
		if segment == "" {
			continue
		}

		currentPath = filepath.Join(currentPath, segment)
		configPath := filepath.Join(currentPath, ".env")

		segmentConfig, err := loadConfig(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return configs, fmt.Errorf("failed to load config at %s: %w", configPath, err)
			}
		}

		maps.Copy(configs, segmentConfig)
	}

	return configs, nil
}

func (ctx *RqContext) setDockRoot() {
	root, err := ctx.GetDockRoot()
	if err != nil {
		fmt.Printf("Error: %s is not a valid RQ environment\n", ctx.Path)
		os.Exit(1)
	}
	ctx.Dock = root
}

func GetContext() *RqContext {
	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	path = filepath.Clean(path)
	ctx := &RqContext{Path: path, Dock: ""}
	ctx.setDockRoot()

	return ctx
}
func (ctx *RqContext) GetConfigForEnv(relpath, env string) (map[string]string, error) {
	configs := make(map[string]string)

	baseConfig, err := ctx.GetConfig(relpath)
	if err != nil {
		return configs, err
	}
	maps.Copy(configs, baseConfig)
	if env != "" {
		envConfigPath := filepath.Join(ctx.Dock, relpath, ".env."+env)
		envConfig, err := loadConfig(envConfigPath)
		if err != nil && !os.IsNotExist(err) {
			return configs, fmt.Errorf("failed to load environment config %s: %w", envConfigPath, err)
		}
		maps.Copy(configs, envConfig)
	}

	return configs, nil
}
