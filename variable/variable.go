// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package variable

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type VariableContext struct {
	block string
	args  []string
}

type VariableResolver struct {
	env       map[string]string
	functions map[string]func(...string) (string, error)
	re        *regexp.Regexp
}

func NewVariableResolver(env map[string]string) *VariableResolver {
	resolver := &VariableResolver{
		env:       env,
		re:        regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`),
		functions: make(map[string]func(...string) (string, error)),
	}

	resolver.RegisterFunc("uuid", generateUUID)
	resolver.RegisterFunc("file", getFile)
	resolver.RegisterFunc("sha256", generateSHA256)
	resolver.RegisterFunc("timestamp", getCurrentTimestamp)
	resolver.RegisterFunc("now", getNow)
	resolver.RegisterFunc("base64", generateBase64)
	resolver.RegisterFunc("join", joinArgs)

	return resolver
}

func (resolver *VariableResolver) Resolve(value string) (string, error) {
	matches := resolver.re.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		if len(match) > 1 {
			expression := strings.TrimSpace(match[1])
			if expression == "" {
				return "", fmt.Errorf("empty variable expression")
			}

			_, err := resolver.evaluateExpression(expression)
			if err != nil {
				return "", fmt.Errorf("error in expression '{{%s}}': %w", expression, err)
			}
		}
	}

	result := resolver.re.ReplaceAllStringFunc(value, func(match string) string {
		submatches := resolver.re.FindStringSubmatch(match)
		if len(submatches) > 1 {
			expression := strings.TrimSpace(submatches[1])
			val, err := resolver.evaluateExpression(expression)
			if err != nil {
				return match
			}
			return val
		}
		return match
	})

	return result, nil
}

func (resolver *VariableResolver) ResolveFile(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", path)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return resolver.Resolve(string(file))
}

func (resolver *VariableResolver) evaluateExpression(expression string) (string, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return "", fmt.Errorf("empty expression")
	}

	parenIndex := strings.Index(expression, "(")
	if parenIndex > 0 {
		funcname := strings.TrimSpace(expression[:parenIndex])
		if funcname == "" {
			return "", fmt.Errorf("empty function name")
		}

		params := resolver.getParams(expression[parenIndex+1:])

		for i, param := range params {
			val, err := resolver.evaluateExpression(strings.TrimSpace(param))
			if err != nil {
				return "", fmt.Errorf("error in parameter %d of function %s: %w", i+1, funcname, err)
			}
			params[i] = val
		}

		return resolver.evaluateFunction(funcname, params...)
	}

	if variable, ok := resolver.env[expression]; ok {
		return variable, nil
	} else if isString(expression) {
		return expression[1 : len(expression)-1], nil
	}

	return "", fmt.Errorf("variable '%s' not found", expression)
}

func isString(expression string) bool {
	re := regexp.MustCompile(`^'[^']*'$|^"[^"]*"$`)
	return re.MatchString(expression)
}

func (resolver *VariableResolver) getParams(params string) []string {
	if strings.TrimSpace(params) == "" {
		return []string{}
	}

	var res []string
	depth := 0
	accumulated := ""

	for i := 0; i < len(params); i++ {
		char := params[i]

		if char == ',' && depth == 0 {
			res = append(res, strings.TrimSpace(accumulated))
			accumulated = ""
			continue
		}

		switch char {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				if accumulated != "" {
					res = append(res, strings.TrimSpace(accumulated))
				}
				return res
			}
		}

		accumulated += string(char)
	}

	if accumulated != "" {
		res = append(res, strings.TrimSpace(accumulated))
	}

	return res
}
