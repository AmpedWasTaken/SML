package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// assetStyles is an optional shared, inert presentation stylesheet.
// Classes may set only recognized visual attributes; never action metadata.
type assetStyles struct {
	Tokens  map[string]string            `json:"tokens"`
	Classes map[string]map[string]string `json:"classes"`
	Defaults map[string]map[string]string `json:"defaults"`
}

func loadAssetStyles(dir string) (*assetStyles, error) {
	path := filepath.Join(dir, "styles.json")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxFileSize {
		return nil, errors.New("invalid SML styles.json file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var styles assetStyles
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&styles); err != nil {
		return nil, fmt.Errorf("styles.json: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("styles.json: unexpected trailing data")
	}
	if len(styles.Tokens) > 128 || len(styles.Classes) > 128 || len(styles.Defaults) > 32 {
		return nil, errors.New("styles.json: too many tokens or classes")
	}
	for name, value := range styles.Tokens {
		if !validName.MatchString(name) || len(value) > 64 || strings.HasPrefix(value, "@") {
			return nil, fmt.Errorf("styles.json: invalid token %q", name)
		}
	}
	for name, attrs := range styles.Classes {
		if !validName.MatchString(name) || len(attrs) == 0 {
			return nil, fmt.Errorf("styles.json: invalid class %q", name)
		}
		for key, value := range attrs {
			resolved, err := styles.resolve(key, value)
			if err != nil {
				return nil, fmt.Errorf("styles.json: class %q: %w", name, err)
			}
			attrs[key] = resolved
		}
	}
	for kind, attrs := range styles.Defaults {
		if !styleTargets[kind] || len(attrs) == 0 {
			return nil, fmt.Errorf("styles.json: invalid defaults target %q", kind)
		}
		for key, value := range attrs {
			if props, ok := inlineProps[kind]; ok && !props[key] {
				return nil, fmt.Errorf("styles.json: unsupported %s default %q", kind, key)
			}
			resolved, err := styles.resolve(key, value)
			if err != nil {
				return nil, fmt.Errorf("styles.json: defaults for %q: %w", kind, err)
			}
			attrs[key] = resolved
		}
	}
	return &styles, nil
}

func sharedStyleAttribute(key string) bool {
	return styleAttributes[key] && key != "animate" && key != "duration-ms" && key != "easing"
}

func (s *assetStyles) resolve(key, value string) (string, error) {
	if !sharedStyleAttribute(key) {
		return "", fmt.Errorf("unsupported shared style attribute %q", key)
	}
	if strings.HasPrefix(value, "@") {
		name := strings.TrimPrefix(value, "@")
		var ok bool
		if s != nil {
			value, ok = s.Tokens[name]
		}
		if !ok {
			return "", fmt.Errorf("unknown SML design token %q", name)
		}
	}
	if err := validateStyleAttribute(key, value); err != nil {
		return "", err
	}
	return value, nil
}

// Defaults have lowest priority; classes are applied left to right; explicit inline attributes win.
// The ordinary parser still validates every resolved attribute against its node.
func applyAssetStyles(nodes []Node, styles *assetStyles) error {
	for i := range nodes {
		n := &nodes[i]
		if spec, ok := n.Attrs["class"]; ok {
			if !styleTargets[n.Kind] || styles == nil {
				return fmt.Errorf("SML <%s> class requires a shared stylesheet and a visual element", n.Kind)
			}
			classes := strings.Fields(spec)
			if len(classes) == 0 || len(classes) > 8 {
				return errors.New("SML class requires 1..8 class names")
			}
			merged := map[string]string{}
			for _, name := range classes {
				attrs, ok := styles.Classes[name]
				if !ok {
					return fmt.Errorf("unknown SML style class %q", name)
				}
				for key, value := range attrs {
					merged[key] = value
				}
			}
			for key, value := range n.Attrs {
				if key != "class" {
					merged[key] = value
				}
			}
			n.Attrs = merged
		}
		if styles != nil {
			for key, value := range styles.Defaults[n.Kind] {
				if _, exists := n.Attrs[key]; !exists {
					n.Attrs[key] = value
				}
			}
		}
		for key, value := range n.Attrs {
			if sharedStyleAttribute(key) && strings.HasPrefix(value, "@") {
				resolved, err := styles.resolve(key, value)
				if err != nil {
					return fmt.Errorf("SML <%s>: %w", n.Kind, err)
				}
				n.Attrs[key] = resolved
			}
		}
		if err := applyAssetStyles(n.Children, styles); err != nil {
			return err
		}
	}
	return nil
}
