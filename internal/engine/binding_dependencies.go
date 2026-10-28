package engine

import (
	"sort"
	"strings"
)

// BindingDependencies is a conservative description of host binding keys
// that may affect a parsed SML view. It is computed without terminal I/O,
// running markup, or accessing remote data.
//
// Conservative means that source-linked components or unknown nodes prevent
// a complete static dependency graph, so any changed host binding warrants a
// render. This avoids accidentally preserving stale controls or text.
type BindingDependencies struct {
	Keys         []string
	Conservative bool
}

// AffectedBy reports whether a batch of changed host keys may alter the view.
// A false result is only a hint that a host may skip a binding-driven redraw.
// Hosts must still redraw for resize, animation, assets, viewport movement,
// theme/capability changes and other external presentation changes.
func (d BindingDependencies) AffectedBy(changedKeys ...string) bool {
	if len(changedKeys) == 0 {
		return false
	}
	if d.Conservative {
		return true
	}
	for _, changed := range changedKeys {
		for _, dependency := range d.Keys {
			if changed == dependency {
				return true
			}
		}
	}
	return false
}

// BindingDependencies inspects a view's declarative sources, derived-value
// graph and interpolation sites. Conditionals and hidden branches count as
// dependencies: changes may make controls appear/disappear. Local includes
// and component uses trigger conservative fallback, since their expanded
// declarations cannot be guaranteed from the parent view alone.
//
// The result is detached from the View. It may be calculated once after
// Parse/LoadDir and regenerated whenever the host swaps the view.
func (v View) BindingDependencies() BindingDependencies {
	used := make(map[string]bool)
	derived := make(map[string][]string)
	uncertain := false
	nodesSeen := 0

	add := func(key string) {
		if key != "" {
			used[key] = true
		}
	}
	interpolations := func(s string) {
		for _, match := range binding.FindAllStringSubmatch(s, -1) {
			add(match[1])
		}
	}

	var scan func([]Node, int)
	scan = func(nodes []Node, depth int) {
		if depth > maxDepth {
			uncertain = true
			return
		}
		for _, n := range nodes {
			nodesSeen++
			if nodesSeen > maxNodes {
				uncertain = true
				return
			}
			if !allowed[n.Kind] || n.Kind == "use" || n.Kind == "include" {
				uncertain = true
			}
			if n.Kind == "derive" {
				key := n.Attr("key", "")
				if key == "" {
					uncertain = true
				} else {
					derived[key] = append(derived[key], n.Attr("source", ""))
					if n.Attr("other", "") != "" {
						derived[key] = append(derived[key], n.Attr("other", ""))
					}
				}
			} else {
				interpolations(n.Text)
				for _, attr := range n.Attrs {
					interpolations(attr)
				}
				// These elements read a binding key by attribute name, not
				// by {{interpolation}}. This includes render and focus gates.
				switch n.Kind {
				case "field", "toggle", "choice", "input", "number":
					add(n.Attr("key", ""))
				case "if", "when":
					if field := n.Attr("field", ""); field != "" {
						add(field)
					} else {
						add(n.Attr("value", ""))
					}
				case "empty":
					add(n.Attr("field", ""))
				case "repeat", "table":
					add(n.Attr("source", ""))
				case "prompt", "stat":
					add(n.Attr("value", ""))
				}
				add(n.Attr("enabled-when", ""))
			}
			scan(n.Children, depth+1)
		}
	}
	scan(v.Nodes, 0)

	// Derived declarations are ordered, presentation-only expressions. Walk
	// their sources transitively, but do not mark unused derived sources as
	// necessary to repaint. A visited set prevents cycles in manually built
	// or mutated view trees that did not pass Parse.
	expanded := make(map[string]bool)
	var expand func(string)
	expand = func(key string) {
		if expanded[key] {
			return
		}
		expanded[key] = true
		for _, dependency := range derived[key] {
			add(dependency)
			if strings.HasPrefix(dependency, "derived.") {
				expand(dependency)
			}
		}
	}
	for key := range used {
		if strings.HasPrefix(key, "derived.") {
			expand(key)
		}
	}

	result := BindingDependencies{Conservative: uncertain}
	for key := range used {
		// derived.* is an SML-reserved computed namespace. The caller
		// cannot directly update it through typed host state.
		if !strings.HasPrefix(key, "derived.") {
			result.Keys = append(result.Keys, key)
		}
	}
	sort.Strings(result.Keys)
	return result
}
