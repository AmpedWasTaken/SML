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
	state   string
	parents []cssCompoundRule
	direct  []bool
	filters []cssAttributeFilter
}

type cssAttributeFilter struct {
	key, value string
	hasValue   bool
}

// parseContextSelector supports at most four simple selector segments.
func parseContextSelector(selector string) (cssCompoundRule, error) {
	parts := strings.Fields(strings.ReplaceAll(selector, ">", " > "))
	var segments []cssCompoundRule
	var direct []bool
	nextDirect := false
	for _, part := range parts {
		if part == ">" {
			if len(segments) == 0 || nextDirect {
				return cssCompoundRule{}, fmt.Errorf("invalid child combinator")
			}
			nextDirect = true
			continue
		}
		if len(segments) >= 4 {
			return cssCompoundRule{}, fmt.Errorf("selector chain exceeds four segments")
		}
		base, state, hasState := strings.Cut(part, ":")
		var segment cssCompoundRule
		var err error
		if styleTargets[base] {
			segment.kind = base
		} else {
			segment, err = parseCompoundSelector(base)
			if err != nil {
				return segment, err
			}
		}
		if hasState {
			if _, _, err := cssStateSelector(segment.kind + ":" + state); err != nil {
				return segment, err
			}
			segment.state = state
		}
		if len(segments) > 0 {
			if segments[len(segments)-1].state != "" {
				return segment, fmt.Errorf("pseudo-selector must be on final control")
			}
			direct = append(direct, nextDirect)
		}
		segments = append(segments, segment)
		nextDirect = false
	}
	if len(segments) == 0 || nextDirect {
		return cssCompoundRule{}, fmt.Errorf("invalid contextual selector")
	}
	last := segments[len(segments)-1]
	last.parents, last.direct = segments[:len(segments)-1], direct
	return last, nil
}

func parseCompoundSelector(selector string) (cssCompoundRule, error) {
	rule := cssCompoundRule{}
	rest := selector
	if i := strings.IndexAny(rest, ".#["); i > 0 {
		rule.kind, rest = rest[:i], rest[i:]
		if !styleTargets[rule.kind] {
			return rule, fmt.Errorf("unknown compound selector target %q", rule.kind)
		}
	}
	for rest != "" {
		prefix := rest[0]
		if prefix == '[' {
			end := strings.IndexByte(rest, ']')
			if end < 0 || len(rule.filters) >= 4 {
				return rule, fmt.Errorf("invalid or excessive attribute selectors")
			}
			filter, err := parseCSSAttributeFilter(rest[1:end])
			if err != nil {
				return rule, err
			}
			rule.filters = append(rule.filters, filter)
			rest = rest[end+1:]
			continue
		}
		if prefix != '.' && prefix != '#' {
			return rule, fmt.Errorf("invalid compound selector %q", selector)
		}
		rest = rest[1:]
		end := strings.IndexAny(rest, ".#[")
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
	if rule.id == "" && len(rule.classes) == 0 && len(rule.filters) == 0 {
		return rule, fmt.Errorf("compound selector requires a class, ID or attribute")
	}
	return rule, nil
}

func (r cssCompoundRule) specificity() int {
	score := (len(r.classes) + len(r.filters)) * 100
	if r.state != "" {
		score += 100
	}
	if r.id != "" {
		score += 10000
	}
	if r.kind != "" {
		score++
	}
	for _, parent := range r.parents {
		score += parent.specificity()
	}
	return score
}

func (r cssCompoundRule) matches(n Node, classes []string) bool {
	if !styleTargets[n.Kind] || (r.kind != "" && r.kind != n.Kind) ||
		(r.id != "" && r.id != n.Attr("id", "")) {
		return false
	}
	for _, filter := range r.filters {
		value, exists := n.Attrs[filter.key]
		if !exists || (filter.hasValue && value != filter.value) {
			return false
		}
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

type cssAncestor struct {
	node    Node
	classes []string
}

// Match the source tree, not expanded component call-site ancestry.
// Dynamic programming avoids exponential descendant backtracking.
func (r cssCompoundRule) matchesContext(n Node, classes []string, ancestors []cssAncestor) bool {
	if !r.matches(n, classes) {
		return false
	}
	if len(r.parents) == 0 {
		return true
	}
	chain := append(append([]cssCompoundRule(nil), r.parents...), r)
	nodes := append(append([]cssAncestor(nil), ancestors...), cssAncestor{n, classes})
	var previous []bool
	for i, selector := range chain {
		current := make([]bool, len(nodes))
		seen := false
		for j, item := range nodes {
			eligible := i == 0
			if i > 0 && j > 0 {
				seen = seen || previous[j-1]
				eligible = seen
				if r.direct[i-1] {
					eligible = previous[j-1]
				}
			}
			current[j] = eligible && selector.matches(item.node, item.classes)
		}
		previous = current
	}
	return previous[len(nodes)-1]
}

func validateCompoundRules(styles *assetStyles) error {
	for _, rule := range styles.Compound {
		for key, value := range rule.attr {
			if rule.state != "" && key != "foreground" && key != "background" && key != "border-color" {
				return fmt.Errorf("contextual state cannot change %s", key)
			}
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
			if rule.state != "" && resolved == "inherit" {
				return fmt.Errorf("state selectors require explicit colors")
			}
			rule.attr[key] = resolved
		}
	}
	return nil
}

// Keep the legacy class-order cascade. Compound selectors are a separate,
// higher-specificity layer; explicit markup remains the final override.
func (s *assetStyles) compoundAttrs(n Node, classes []string, ancestors []cssAncestor, merged map[string]string) error {
	scores := make(map[string]int)
	for _, rule := range s.Compound {
		if rule.state != "" || !rule.matchesContext(n, classes, ancestors) {
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
		for _, segment := range append(append([]cssCompoundRule(nil), rule.parents...), rule) {
			for _, class := range segment.classes {
				if name == class {
					return true
				}
			}
		}
	}
	return false
}

func (s *assetStyles) contextualState(n Node, classes []string, ancestors []cssAncestor, state string) cssStatePaint {
	attrs := make(map[string]string)
	for key, value := range s.States[n.Kind+":"+state] {
		attrs[key] = value
	}
	scores := make(map[string]int)
	for _, rule := range s.Compound {
		if rule.state != state || !rule.matchesContext(n, classes, ancestors) {
			continue
		}
		for key, value := range rule.attr {
			if score, ok := scores[key]; !ok || rule.specificity() >= score {
				attrs[key] = value
				scores[key] = rule.specificity()
			}
		}
	}
	return cssStatePaint{Foreground: attrs["foreground"], Background: attrs["background"], BorderColor: attrs["border-color"]}
}

func parseCSSAttributeFilter(source string) (cssAttributeFilter, error) {
	key, value, hasValue := strings.Cut(source, "=")
	// Match only literal visual attributes, not action metadata or bindings.
	switch key {
	case "variant", "tone", "appearance", "border", "align", "case":
	default:
		return cssAttributeFilter{}, fmt.Errorf("unsupported attribute selector %q", key)
	}
	if hasValue {
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if !validName.MatchString(value) || validateStyleAttribute(key, value) != nil {
			return cssAttributeFilter{}, fmt.Errorf("invalid attribute selector value %q", value)
		}
	}
	return cssAttributeFilter{key: key, value: value, hasValue: hasValue}, nil
}
