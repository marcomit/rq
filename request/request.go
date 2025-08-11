// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import "fmt"

func parseRequest() {
	if len(args) == 0 {
		fmt.Println("Invalid arguments")
		return
	}

	switch args[0] {
	case "new":

	case "run":
	}
}

func createRequest(path string) {

}

func runRequest(path string) {

}
