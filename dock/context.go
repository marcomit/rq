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

type RqContext struct {
	Path []string
	Dock string
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func validatePath(path string, predicate func(string) bool) []string {
	sep := string(os.PathSeparator)
	paths := strings.Split(path, sep)

	length := len(paths)

	res := []string{}

	for i := length; i > 0; i-- {
		curr := strings.Join(paths[0:i], sep)
		if predicate(curr) {
			res = append(res, curr)
		}
	}

	return res
}

func (ctx *RqContext) GetPath() string {
	return filepath.Join(ctx.Path...)
}

func (ctx *RqContext) IsValidDock() bool {

	res := validatePath(ctx.GetPath(), func(curr string) bool {
		return exists(filepath.Join(curr, ".dock"))
	})

	return len(res) > 0
}

func (ctx *RqContext) GetDockRoot() (string, error) {

	res := validatePath(ctx.GetPath(), func(curr string) bool {
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

func GetContext() *RqContext {
	path, _ := os.Getwd()

	ctx := &RqContext{strings.Split(path, string(os.PathSeparator)), ""}
	if !ctx.IsValidDock() {
		fmt.Println(path, "is not a valid RQ environment")
	}

	ctx.Dock, _ = ctx.GetDockRoot()

	return ctx
}
