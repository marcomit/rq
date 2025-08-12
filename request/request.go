// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"fmt"
	"os"
	"path/filepath"
	"rq/dock"
	"rq/variable"
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
		Evaluate(ctx, path)
	default:
		fmt.Println("Multiple requests detected:")
		for i := range requests {
			fmt.Println(i, ".", requests[i])
		}
	}
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
func EvaluateWithOptions(ctx *dock.RqContext, request string, options ExecuteOptions) error {
	requestPath := filepath.Join(ctx.Dock, request)
	if !strings.HasSuffix(requestPath, ".http") {
		requestPath += ".http"
	}

	var config map[string]string
	var err error
	if options.Environment != "" {
		config, err = ctx.GetConfigForEnv(filepath.Dir(request), options.Environment)
	} else {
		config, err = ctx.GetConfig(filepath.Dir(request))
	}
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	resolver := variable.NewVariableResolver(config)
	content, err := resolver.ResolveFile(requestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve variables in request: %w", err)
	}

	httpReq, err := ParseHttpRequest(content)
	if err != nil {
		return fmt.Errorf("failed to parse HTTP request: %w", err)
	}

	fmt.Printf("Executing %s %s\n", httpReq.Method, httpReq.URL)
	response, err := httpReq.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	if options.OutputFile != "" {
		if options.OutputBodyOnly {
			err = os.WriteFile(options.OutputFile, []byte(response.Body), 0644)
		} else {
			err = response.SaveToFile(options.OutputFile)
		}
		if err != nil {
			return fmt.Errorf("failed to save output: %w", err)
		}
		fmt.Printf("Response saved to %s\n", options.OutputFile)
	} else {
		response.Print()
	}

	return nil
}
