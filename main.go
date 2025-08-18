// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"os"
	"rq/dock"
	"rq/docs"
	"rq/environment"
	"rq/request"

	"github.com/marcomit/args"
)

func main() {
	rq := args.New("rq").Action(func(r *args.Result) error {
		fmt.Println("Welcome to RQ!")
		return nil
	})

	dock.Setup(rq)
	request.Setup(rq)
	environment.Setup(rq)
	docs.Setup(rq)

	err := rq.Run(os.Args[1:])

	if err != nil {
		fmt.Println(err)
	}

	if len(os.Args) == 1 {
		rq.Usage()
	}
}
