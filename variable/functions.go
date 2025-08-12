package variable

import (
	"fmt"
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
	return "", fmt.Errorf("Function does not exists")
}

func getFile(args ...string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

func generateUUID(args ...string) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("You cannot pass arguments to the uuid function")
	}
	return uuid.New().String(), nil
}
