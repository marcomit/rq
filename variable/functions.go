// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package variable

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (resolver *VariableResolver) RegisterFunc(funcname string, callback func(...string) (string, error)) error {
	if _, ok := resolver.functions[funcname]; ok {
		return fmt.Errorf("%v already registered", funcname)
	}
	resolver.functions[funcname] = callback
	return nil
}

func (resolver *VariableResolver) ContainsFunc(funcname string) bool {
	_, ok := resolver.functions[funcname]
	return ok
}

func (resolver *VariableResolver) evaluateFunction(funcname string, params ...string) (string, error) {
	if fn, ok := resolver.functions[funcname]; ok {
		return fn(params...)
	}
	return "", fmt.Errorf("function %s does not exist", funcname)
}

func getFile(args ...string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("file() function expects exactly 1 argument, got %d", len(args))
	}

	filepath := args[0]
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filepath)
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

func generateUUID(args ...string) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("uuid() function takes no arguments, got %d", len(args))
	}
	return uuid.New().String(), nil
}

func generateSHA256(args ...string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("sha256() function expects exactly 1 argument, got %d", len(args))
	}

	input := args[0]

	if _, err := os.Stat(input); err == nil {
		file, err := os.Open(input)
		if err != nil {
			return "", fmt.Errorf("failed to open file %s: %w", input, err)
		}
		defer file.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return "", fmt.Errorf("failed to hash file %s: %w", input, err)
		}
		return fmt.Sprintf("%x", hasher.Sum(nil)), nil
	}

	hasher := sha256.New()
	hasher.Write([]byte(input))
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func getCurrentTimestamp(args ...string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("timestamp() function expects 0 or 1 argument, got %d", len(args))
	}

	format := time.RFC3339
	if len(args) == 1 {
		format = args[0]
	}

	return time.Now().Format(format), nil
}

func getNow(args ...string) (string, error) {
	return getCurrentTimestamp(args...)
}

func joinArgs(s ...string) (string, error) {
	if len(s) < 2 {
		return "", fmt.Errorf("join() function expects at least 2 parameters, got %d", len(s))
	}
	separator := s[len(s)-1]
	return strings.Join(s[:len(s)-1], separator), nil
}

func generateBase64(s ...string) (string, error) {
	if len(s) != 1 {
		return "", fmt.Errorf("base64() function expects exactly 1 paramenter, got %d", len(s))
	}
	return base64.StdEncoding.EncodeToString([]byte(s[0])), nil
}
