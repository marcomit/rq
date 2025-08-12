// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package dock

import (
	"fmt"
	"os"
	"path/filepath"
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
		return "", fmt.Errorf("No valid path found")
	}

	return res[0], nil
}

func (ctx *RqContext) GetConfig(path string) map[string]string {
	configs := make(map[string]string)

	return configs
}

func (ctx *RqContext) setDockRoot(path string) {
	root, err := ctx.GetDockRoot()

	if err != nil {
		fmt.Println(ctx.Path, "is not a valid RQ environment")
		os.Exit(1)
	}
	ctx.Path = root
}

func GetContext() *RqContext {
	path, _ := os.Getwd()

	path = filepath.Clean(path)

	ctx := &RqContext{path, ""}

	ctx.setDockRoot(path)

	return ctx
}
