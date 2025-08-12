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
	fmt.Println("Usage: rq dock create <name of dock>")
	args := os.Args[1:]
	parseArgs(args)
}

func parseArgs(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid arguments")
		os.Exit(1)
	}

	fmt.Println(args)
	switch args[0] {
	case "dock":
		fmt.Println("Parsing dock")
		dock.Parse(args[1:])
	case "new":
		ctx := dock.GetContext()
		request.New(ctx, args[1], "http")

	case "run":
		fmt.Println(args)
		ctx := dock.GetContext()
		request.Run(ctx, args[1])
	default:
		fmt.Println("Invalid rq command")
	}
}
