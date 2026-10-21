package sml

import (
	"fmt"
	"regexp"
	"strings"
)

// Bindings supplies runtime values for {{name}} placeholders. Values are plain
// text, not XML or commands; binding happens after parsing and is immutable.
type Bindings map[string]string

var interpolation = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_.-]*)\s*\}\}`)

// Bind creates a deep copy of the SML document with placeholders replaced
// by provided values. The source tree remains unchanged and reusable.
func Bind(root *Node, data Bindings) (*Node, error) {
	if root == nil || root.Tag != "sml" {
		return nil, fmt.Errorf("expected a parsed <sml> root")
	}
	return bindNode(root, data)
}

func bindNode(node *Node, data Bindings) (*Node, error) {
	clone := &Node{Tag: node.Tag, Attrs: make(map[string]string, len(node.Attrs))}
	var err error
	clone.Text, err = replaceValues(node.Text, data)
	if err != nil {
		return nil, err
	}
	for key, value := range node.Attrs {
		clone.Attrs[key], err = replaceValues(value, data)
		if err != nil {
			return nil, fmt.Errorf("attribute %s: %w", key, err)
		}
	}
	for _, child := range node.Children {
		copyChild, err := bindNode(child, data)
		if err != nil {
			return nil, err
		}
		clone.Children = append(clone.Children, copyChild)
	}
	return clone, nil
}

func replaceValues(source string, data Bindings) (string, error) {
	var missing string
	result := interpolation.ReplaceAllStringFunc(source, func(match string) string {
		key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}"))
		value, exists := data[key]
		if !exists {
			missing = key
			return match
		}
		if len(value) > 4096 {
			missing = key + " (value exceeds 4096 bytes)"
			return match
		}
		return value
	})
	if missing != "" {
		return "", fmt.Errorf("missing or oversized SML binding: %s", missing)
	}
	return result, nil
}

// RenderData parses and renders a reusable document against live application values.
func RenderData(root *Node, data Bindings, options Options) (string, error) {
	bound, err := Bind(root, data)
	if err != nil {
		return "", err
	}
	return RenderString(bound, options)
}
