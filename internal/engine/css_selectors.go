package engine

import (
	"fmt"
	"strings"
)

type cssCompoundRule struct {
	kind    string
	id      string
	classes []string
	attr    map[string]string
}

func parseCompoundSelector(selector string) (cssCompoundRule, error) {
	rule := cssCompoundRule{}
	rest := selector
	if i := strings.IndexAny(rest, ".#"); i > 0 {
		rule.kind, rest = rest[:i], rest[i:]
		if !styleTargets[rule.kind] {
			return rule, fmt.Errorf("unknown compound selector target %q", rule.kind)
		}
	}
	for rest != "" {
		prefix := rest[0]
		if prefix != '.' && prefix != '#' {
			return rule, fmt.Errorf("invalid compound selector %q", selector)
		}
		rest = rest[1:]
		end := strings.IndexAny(rest, ".#")
		if end < 0 {
			end = len(rest)
		}
		name := rest[:end]
		rest = rest[end:]
		if !validName.MatchString(name) {
			return rule, fmt.Errorf("invalid selector name %q", name)
		}
		if prefix == '#' {
			if rule.id != "" {
				return rule, fmt.Errorf("only one ID per selector is supported")
			}
			rule.id = name
			continue
		}
		for _, existing := range rule.classes {
			if existing == name {
				return rule, fmt.Errorf("duplicate selector class %q", name)
			}
		}
		rule.classes = append(rule.classes, name)
		if len(rule.classes) > 8 {
			return rule, fmt.Errorf("too many selector classes")
		}
	}
	if rule.id == "" && len(rule.classes) == 0 {
		return rule, fmt.Errorf("compound selector requires a class or ID")
	}
	return rule, nil
}

func (r cssCompoundRule) specificity() int {
	score := len(r.classes) * 10
	if r.id != "" {
		score += 100
	}
	if r.kind != "" {
		score++
	}
	return score
}

func (r cssCompoundRule) matches(n Node, classes []string) bool {
	if !styleTargets[n.Kind] || (r.kind != "" && r.kind != n.Kind) ||
		(r.id != "" && r.id != n.Attr("id", "")) {
		return false
	}
	for _, name := range r.classes {
		found := false
		for _, actual := range classes {
			if name == actual {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func validateCompoundRules(styles *assetStyles) error {
	for _, rule := range styles.Compound {
		for key, value := range rule.attr {
			if rule.kind != "" {
				if layoutStyleAttribute(key) && !layoutStyleTarget(rule.kind, key) {
					return fmt.Errorf("unsupported compound layout property %q on %s", key, rule.kind)
				}
				if props, ok := inlineProps[rule.kind]; ok && !props[key] {
					return fmt.Errorf("unsupported compound property %q on %s", key, rule.kind)
				}
			}
			resolved, err := styles.resolve(key, value)
			if err != nil {
				return err
			}
			rule.attr[key] = resolved
		}
	}
	return nil
}

// Keep the legacy class-order cascade. Compound selectors are a separate,
// higher-specificity layer; explicit markup remains the final override.
func (s *assetStyles) compoundAttrs(n Node, classes []string, merged map[string]string) error {
	scores := make(map[string]int)
	for _, rule := range s.Compound {
		if !rule.matches(n, classes) {
			continue
		}
		score := rule.specificity()
		for key, value := range rule.attr {
			if layoutStyleAttribute(key) && !layoutStyleTarget(n.Kind, key) {
				return fmt.Errorf("unsupported compound layout property %q on %s", key, n.Kind)
			}
			if previous, ok := scores[key]; !ok || score >= previous {
				merged[key] = value
				scores[key] = score
			}
		}
	}
	return nil
}

func (s *assetStyles) hasCompoundClass(name string) bool {
	for _, rule := range s.Compound {
		for _, class := range rule.classes {
			if name == class {
				return true
			}
		}
	}
	return false
}
