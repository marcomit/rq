// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package dock

import (
	"fmt"
	"os"
	"path/filepath"
)

func Parse(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid dock parameter")
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		if len(args) < 2 {
			fmt.Println("Expected dock name")
			break
		}
		createDock(args[1])
	case "use":
		setCurrentDock(args[1])

	case "list":
		break

	default:
		fmt.Println("Default: Invalid dock command")
	}
}

func setCurrentDock(name string) {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error setting the default dock")
		return
	}
	file, e := os.Create(filepath.Join(dir))
	if e != nil {
		return
	}
	file.WriteString(name)
}

func createDock(name string) {
	fmt.Println("Creating dock", name)
	err := os.Mkdir(name, 0755)
	if err != nil {
		return
	}

	dock, err := os.Create(filepath.Join(name, ".dock"))

	if err != nil {
		return
	}

	_, writeErr := dock.WriteString(name)

	if writeErr != nil {
		fmt.Println("Error saving the dock")
		return
	}

	_, envErr := os.Create(filepath.Join(name, ".env"))

	if envErr != nil {
		fmt.Println("Error saving the default env")
		return
	}
}
