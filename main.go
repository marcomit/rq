// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"os"
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

	switch args[0] {
	case "dock":
		break
	case "new":

	default:
		fmt.Println("Invalid rq command")
	}
}
