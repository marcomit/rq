// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"fmt"
	"os"
	"path/filepath"
	"rq/dock"
	"strings"
)

func New(ctx *dock.RqContext, file string, protocol string) error {
	wd := ctx.Path

	f, err := os.Create(filepath.Join(wd, file+".http"))

	if err != nil {
		return err
	}

	f.WriteString(`GET {{BASE_URL}}/api {{HTTP_VERSION}}
`)

	return nil
}

func Run(ctx *dock.RqContext, path string) {
	fmt.Println("Path ", path)
	requests, err := retrieveRequests(ctx.Path, path)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	switch len(requests) {
	case 0:
		fmt.Println("Request", path, "not found")
	case 1:
		fmt.Println("Run request")
		break
	default:
		fmt.Println("Multiple requests detected:")
		for i := range requests {
			fmt.Println(i, ".", requests[i])
		}
	}
}

func Evaluate() {

}

func retrieveRequests(path string, req string) ([]string, error) {
	entries, err := os.ReadDir(path)
	sep := string(os.PathSeparator)
	if err != nil {
		return []string{}, err
	}

	res := []string{}

	reqpath := strings.Split(req, sep)

	req = reqpath[0]

	for _, entry := range entries {
		if len(reqpath) == 1 && !entry.IsDir() {
			continue
		}

		fileinfo := strings.Split(entry.Name(), ".")
		if fileinfo[0] == req {

			if len(reqpath) == 1 {
				res = append(res, entry.Name())
			} else {
				subpath := filepath.Join(path, entry.Name())

				reqs, e := retrieveRequests(subpath, strings.Join(reqpath[1:], sep))

				if e != nil {
					return res, e
				}

				res = append(res, reqs...)
			}
		}
	}
	return res, nil
}
