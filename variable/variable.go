// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package variable

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type VariableResolver struct {
	env       map[string]string
	functions map[string]func(...string) (string, error)
	re        *regexp.Regexp
}

func NewVariableResolver(env map[string]string) *VariableResolver {
	resolver := &VariableResolver{env: env,
		re:        regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`),
		functions: make(map[string]func(...string) (string, error)),
	}

	resolver.RegisterFunc("uuid", func(s ...string) (string, error) {
		return uuid.NewString(), nil
	})

	resolver.RegisterFunc("join", func(s ...string) (string, error) {
		l := len(s)
		if len(s) < 2 {
			return "", fmt.Errorf("Expected at least 2 parameters")
		}
		separator := s[l-1]
		return strings.Join(s[:l-1], separator), nil
	})

	resolver.RegisterFunc("file", func(s ...string) (string, error) {
		if len(s) != 1 {
			return "", fmt.Errorf("Expected 1 arguments, given %v", len(s))
		}
		return "", nil
	})

	resolver.RegisterFunc("now", func(s ...string) (string, error) {
		if len(s) != 0 {
			return "", fmt.Errorf("Expected zero arguments, given %v", len(s))
		}
		return "", nil
	})

	return resolver
}

func (resolver *VariableResolver) Resolve(value string) (string, error) {

	matches := resolver.re.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		if len(match) > 1 {
			expression := strings.TrimSpace(match[1])

			_, err := resolver.evaluateExpression(strings.TrimSpace(expression))

			if err != nil {
				return "", err
			}
		}
	}

	result := resolver.re.ReplaceAllStringFunc(value, func(match string) string {
		submatches := resolver.re.FindStringSubmatch(match)

		res := match

		if len(submatches) > 1 {
			res = submatches[1]
		}

		val, _ := resolver.evaluateExpression(strings.TrimSpace(res))

		return val
	})

	return result, nil
}

func (resolver *VariableResolver) ResolveFile(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return resolver.Resolve(string(file))
}

func (resolver *VariableResolver) evaluateExpression(expression string) (string, error) {
	var funcname string = ""
	expression = strings.TrimSpace(expression)
	for i := 0; i < len(expression); i++ {
		if expression[i] == '(' {
			params := resolver.getParams(expression[i+1:])
			for j, p := range params {
				val, e := resolver.evaluateExpression(strings.TrimSpace(p))
				if e != nil {
					return "", e
				}
				params[j] = val
			}
			return resolver.evaluateFunction(funcname, params...)

		}

		funcname += string(expression[i])
	}

	if funcname == "" {
		return "", fmt.Errorf("Empty expression")
	}

	if variable, ok := resolver.env[funcname]; ok {
		return variable, nil
	}
	return "", fmt.Errorf("Variable %s not found", funcname)
}

func (resolver *VariableResolver) getParams(params string) []string {
	var res []string
	depth := 0
	accumulated := ""

	for i := 0; i < len(params); i++ {
		if params[i] == ',' && depth == 0 {
			res = append(res, accumulated)
			accumulated = ""
			continue
		}
		switch params[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				res = append(res, accumulated)
				return res
			}
		}
		accumulated += string(params[i])
	}

	if accumulated != "" {
		res = append(res, accumulated)
	}

	for i := range res {
		res[i] = strings.TrimSpace(res[i])
	}

	return res
}
