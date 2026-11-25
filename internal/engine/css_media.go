package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// Media rules use the host's full terminal width, not nested pane widths.
// Rules are evaluated without mutating the published asset snapshot.
type cssMediaRule struct {
	kind string
	min  int
	max  int
	attr map[string]string
}

func parseCSSMedia(selector, body string) ([]cssMediaRule, error) {
	condition := strings.TrimSpace(strings.TrimPrefix(selector, "@media"))
	if !strings.HasPrefix(condition, "(") || !strings.HasSuffix(condition, ")") {
		return nil, fmt.Errorf("SMLCSS media requires (min-width: Nch) or (max-width: Nch)")
	}
	key, raw, ok := strings.Cut(condition[1:len(condition)-1], ":")
	key = strings.TrimSpace(key)
	width, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(raw), "ch"))
	if !ok || err != nil || width < 12 || width > 240 {
		return nil, fmt.Errorf("SMLCSS media width must be 12..240 terminal cells")
	}
	low, high := 12, 240
	switch key {
	case "min-width":
		low = width
	case "max-width":
		high = width
	default:
		return nil, fmt.Errorf("unsupported SMLCSS media condition %q", key)
	}
	var rules []cssMediaRule
	for strings.TrimSpace(body) != "" {
		kind, declarations, tail, err := cssBlock(body)
		if err != nil {
			return nil, err
		}
		body = tail
		if kind != "box" && kind != "grid" && kind != "row" {
			return nil, fmt.Errorf("media rules support box, grid and row selectors")
		}
		attrs, err := cssDeclarations(declarations)
		if err != nil {
			return nil, err
		}
		rules = append(rules, cssMediaRule{kind: kind, min: low, max: high, attr: attrs})
		if len(rules) > 32 {
			return nil, fmt.Errorf("too many SMLCSS media selectors")
		}
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("empty SMLCSS media rule")
	}
	return rules, nil
}

func validateCSSMedia(styles *assetStyles) error {
	for _, rule := range styles.Media {
		for key, value := range rule.attr {
			if !layoutStyleAttribute(key) || !layoutStyleTarget(rule.kind, key) {
				return fmt.Errorf("unsupported media property %s on %s", key, rule.kind)
			}
			resolved, err := styles.resolve(key, value)
			if err != nil {
				return err
			}
			rule.attr[key] = resolved
		}
		if rule.kind == "grid" {
			columns, hasColumns := rule.attr["columns"]
			weights, hasWeights := rule.attr["weights"]
			if hasColumns != hasWeights {
				return fmt.Errorf("media grid tracks require grid-template-columns")
			}
			if hasColumns {
				count, _ := strconv.Atoi(columns)
				if len(strings.Split(weights, ",")) != count {
					return fmt.Errorf("media grid tracks do not match column count")
				}
			}
		}
	}
	return nil
}

func (v View) responsiveNode(n Node) Node {
	if v.styles == nil || v.cssWidth == 0 {
		return n
	}
	copied := false
	for _, rule := range v.styles.Media {
		if rule.kind != n.Kind || v.cssWidth < rule.min || v.cssWidth > rule.max {
			continue
		}
		for key, value := range rule.attr {
			if n.cssFixed[key] || ((key == "columns" || key == "weights") && (n.cssFixed["columns"] || n.cssFixed["weights"])) {
				continue
			}
			if !copied {
				attrs := make(map[string]string, len(n.Attrs)+len(rule.attr))
				for k, val := range n.Attrs {
					attrs[k] = val
				}
				n.Attrs = attrs
				copied = true
			}
			n.Attrs[key] = value
		}
	}
	return n
}
