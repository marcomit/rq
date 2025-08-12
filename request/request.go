// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"fmt"
	"os"
	"rq/dock"
	"strings"
)

func Parse(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid arguments")
		return
	}

	switch args[0] {
	case "new":

	case "run":
		ctx := dock.GetContext()
		run(ctx, args)
	}
}

func run(ctx *dock.RqContext, args []string) {
	requests, err := retrieveRequests(ctx.GetPath(), args[1])
	if err != nil {
		fmt.Println("Error")
		os.Exit(1)
	}
	switch len(requests) {
	case 0:
		fmt.Println("Request", args[1], "not found")
	case 1:
		// Do the request
		break
	default:
		fmt.Println("Multiple requests detected:")
		for i := range requests {
			fmt.Println(i, ".", requests[i])
		}
	}
}

func retrieveRequests(path string, req string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return []string{}, err
	}

	res := []string{}

	for i := range entries {
		if entries[i].IsDir() {
			continue
		}
		fileinfo := strings.Split(entries[i].Name(), ".")
		if fileinfo[i] == req {
			res = append(res, entries[i].Name())
		}
	}
	return res, nil
}
