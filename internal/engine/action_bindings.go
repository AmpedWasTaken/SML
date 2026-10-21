package engine

import (
	"errors"
	"strings"
)

// ParseActionBindings accepts only explicit argument:state.key pairs.
// Parsed bindings are schema metadata, never executable expressions.
func ParseActionBindings(spec string) (map[string]string, error) {
	if spec == "" || len(spec) > 300 {
		return nil, errors.New("invalid plugin form bindings")
	}
	pairs := strings.Split(spec, ",")
	if len(pairs) < 1 || len(pairs) > 4 {
		return nil, errors.New("plugin form requires 1..4 bound fields")
	}
	result := make(map[string]string, len(pairs))
	usedSource := map[string]bool{}
	for _, pair := range pairs {
		arg, source, ok := strings.Cut(pair, ":")
		if !ok || !sessionArgument.MatchString(arg) || !sessionStateSource.MatchString(source) {
			return nil, errors.New("invalid plugin form field binding")
		}
		if _, exists := result[arg]; exists || usedSource[source] {
			return nil, errors.New("duplicate plugin form binding")
		}
		result[arg] = source
		usedSource[source] = true
	}
	return result, nil
}
