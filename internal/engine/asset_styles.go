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
	Compound    []cssCompoundRule            `json:"-"`
	Transitions map[string]cssTransition     `json:"-"`
	States      map[string]map[string]string `json:"-"`
	Media       []cssMediaRule               `json:"-"`
	Tokens      map[string]string            `json:"tokens"`
	Classes     map[string]map[string]string `json:"classes"`
	Defaults    map[string]map[string]string `json:"defaults"`
}

func loadAssetStyles(dir string) (*assetStyles, error) {
	jsonPath := filepath.Join(dir, "styles.json")
	cssPath := filepath.Join(dir, "styles.smlcss")
	jsonData, err := readStyleAsset(jsonPath)
	if err != nil {
		return nil, err
	}
	cssData, err := readStyleAsset(cssPath)
	if err != nil {
		return nil, err
	}
	if jsonData != nil && cssData != nil {
		return nil, errors.New("use either styles.json or styles.smlcss, not both")
	}
	if jsonData == nil && cssData == nil {
		return nil, nil
	}
	var styles assetStyles
	if cssData != nil {
		parsed, err := parseStyleCSS(string(cssData))
		if err != nil {
			return nil, err
		}
		styles = parsed
	} else {
		decoder := json.NewDecoder(bytes.NewReader(jsonData))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&styles); err != nil {
			return nil, fmt.Errorf("styles.json: %w", err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return nil, errors.New("styles.json: unexpected trailing data")
		}
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
			if layoutStyleAttribute(key) && !layoutStyleTarget(kind, key) {
				return nil, fmt.Errorf("unsupported %s layout default %q", kind, key)
			}
			if props, ok := inlineProps[kind]; ok && !props[key] {
				return nil, fmt.Errorf("styles.json: unsupported %s default %q", kind, key)
			}
			if kind == "table" && key == "style" {
				if !tableStyles[value] {
					return nil, fmt.Errorf("invalid table style %q", value)
				}
				continue
			}
			resolved, err := styles.resolve(key, value)
			if err != nil {
				return nil, fmt.Errorf("styles.json: defaults for %q: %w", kind, err)
			}
			attrs[key] = resolved
		}
	}
	if err := validateCompoundRules(&styles); err != nil {
		return nil, err
	}
	if err := validateCSSMedia(&styles); err != nil {
		return nil, err
	}
	if err := validateCSSStates(&styles); err != nil {
		return nil, err
	}
	return &styles, nil
}

func readStyleAsset(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxFileSize {
		return nil, fmt.Errorf("invalid SML style file %s", filepath.Base(path))
	}
	return os.ReadFile(path)
}

func sharedStyleAttribute(key string) bool {
	return styleAttributes[key] || layoutStyleAttribute(key)
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
	if value == "inherit" && inheritableStyle(key) {
		return value, nil
	}
	if layoutStyleAttribute(key) {
		if err := validateLayoutStyleValue(key, value); err != nil {
			return "", err
		}
	} else if err := validateStyleAttribute(key, value); err != nil {
		return "", err
	}
	return value, nil
}

// Defaults have lowest priority; classes are applied left to right; explicit inline attributes win.
// The ordinary parser still validates every resolved attribute against its node.
func applyAssetStyles(nodes []Node, styles *assetStyles) error {
	return applyContextStyles(nodes, styles, nil, nil)
}

func applyContextStyles(nodes []Node, styles *assetStyles, ancestors []cssAncestor, inherited map[string]string) error {
	for i := range nodes {
		n := &nodes[i]
		spec, hasClass := n.Attrs["class"]
		classes := strings.Fields(spec)
		original := cssAncestor{node: *n, classes: classes}
		if styles != nil && len(styles.Compound) > 0 && controlNode(*n) {
			n.cssPaint = &cssControlPaint{
				Hover:    styles.contextualState(*n, classes, ancestors, "hover"),
				Focus:    styles.contextualState(*n, classes, ancestors, "focus"),
				Disabled: styles.contextualState(*n, classes, ancestors, "disabled"),
			}
		}
		if hasClass || (styles != nil && len(styles.Compound) > 0 && styleTargets[n.Kind]) {
			if !styleTargets[n.Kind] || styles == nil {
				return fmt.Errorf("SML <%s> class requires a shared stylesheet and a visual element", n.Kind)
			}
			if (hasClass && len(classes) == 0) || len(classes) > 8 {
				return errors.New("SML class requires 1..8 class names")
			}
			merged := map[string]string{}
			for _, name := range classes {
				attrs, ok := styles.Classes[name]
				if !ok && !styles.hasCompoundClass(name) {
					return fmt.Errorf("unknown SML style class %q", name)
				}
				for key, value := range attrs {
					if layoutStyleAttribute(key) && !layoutStyleTarget(n.Kind, key) {
						return fmt.Errorf("layout property %q not supported on %s", key, n.Kind)
					}
					merged[key] = value
				}
			}
			if err := styles.compoundAttrs(*n, classes, ancestors, merged); err != nil {
				return err
			}
			for key, value := range n.Attrs {
				if key != "class" {
					merged[key] = value
				}
			}
			n.Attrs = merged
		}
		if styles != nil {
			if len(styles.Media) > 0 && (n.Kind == "box" || n.Kind == "grid" || n.Kind == "row") {
				n.cssFixed = make(map[string]bool, len(n.Attrs))
				for key := range n.Attrs {
					n.cssFixed[key] = true
				}
			}
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
		computed := inherited
		for _, key := range []string{"foreground", "background"} {
			if n.Attr(key, "") == "inherit" {
				if !supportsInheritedStyle(*n, key) {
					return fmt.Errorf("inherit is not supported for %s on %s", key, n.Kind)
				}
				if value := inherited[key]; value != "" {
					n.Attrs[key] = value
				} else {
					delete(n.Attrs, key)
				}
			}
			if value := n.Attr(key, ""); supportsInheritedStyle(*n, key) && validHexColor(value) {
				if computed == nil || computed[key] != value {
					next := make(map[string]string, 2)
					for k, v := range computed {
						next[k] = v
					}
					next[key] = value
					computed = next
				}
			}
		}
		if err := applyContextStyles(n.Children, styles, append(ancestors, original), computed); err != nil {
			return err
		}
	}
	return nil
}

func inheritableStyle(key string) bool {
	return key == "foreground" || key == "background"
}

func supportsInheritedStyle(n Node, key string) bool {
	switch n.Kind {
	case "box", "panel", "button", "text":
		return inheritableStyle(key)
	default:
		return false
	}
}
