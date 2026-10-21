package sml

import (
	"errors"
	"fmt"
	"strings"
)

// Component props are explicit, bounded strings. SML never executes code
// or resolves a component name from a user-controlled data binding.
func validateComponentSyntax(nodes []Node) error {
	var walk func([]Node, string) error
	walk = func(parts []Node, parent string) error {
		for _, n := range parts {
			if id, ok := n.Attrs["id"]; ok && (!controlNode(n) || !actionID.MatchString(id)) {
				return errors.New("invalid SML control id")
			}
			switch n.Kind {
			case "component":
				if !validName.MatchString(n.Attr("name", "")) {
					return errors.New("component needs a safe name")
				}
				if _, err := componentPropNames(n.Attr("props", "")); err != nil {
					return err
				}
				for key := range n.Attrs {
					if key != "name" && key != "props" {
						return fmt.Errorf("unknown component attribute %q", key)
					}
				}
			case "use":
				if !validName.MatchString(n.Attr("name", "")) {
					return errors.New("use needs a safe name")
				}
				for key, value := range n.Attrs {
					if key != "name" && (!validName.MatchString(key) || len(value) > 1024) {
						return errors.New("invalid component prop")
					}
				}
				slots := map[string]bool{}
				for _, child := range n.Children {
					if child.Kind != "fill" {
						return errors.New("use accepts only fill children")
					}
					name := child.Attr("slot", "")
					if !validName.MatchString(name) || slots[name] {
						return errors.New("invalid or duplicate fill slot")
					}
					slots[name] = true
				}
			case "fill":
				if parent != "use" || !validName.MatchString(n.Attr("slot", "")) || len(n.Attrs) != 1 {
					return errors.New("fill requires a named slot in use")
				}
			case "slot":
				if !validName.MatchString(n.Attr("name", "")) || len(n.Attrs) != 1 {
					return errors.New("slot requires a safe name")
				}
			}
			if err := walk(n.Children, n.Kind); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(nodes, "")
}

func componentPropNames(spec string) ([]string, error) {
	if spec == "" {
		return nil, nil
	}
	fields := strings.Split(spec, ",")
	if len(fields) > 24 {
		return nil, errors.New("too many component props")
	}
	names := make([]string, 0, len(fields))
	seen := map[string]bool{}
	for _, field := range fields {
		parts := strings.Split(strings.TrimSpace(field), ":")
		if len(parts) > 2 || (len(parts) == 2 && parts[1] != "string") {
			return nil, errors.New("unsupported prop type")
		}
		name := parts[0]
		if !validName.MatchString(name) || seen[name] {
			return nil, errors.New("invalid or duplicated prop")
		}
		names = append(names, name)
		seen[name] = true
	}
	return names, nil
}

func replaceSlots(nodes []Node, fills map[string][]Node) []Node {
	out := make([]Node, 0, len(nodes))
	for _, n := range nodes {
		if n.Kind == "slot" {
			if assigned, ok := fills[n.Attr("name", "")]; ok {
				out = append(out, assigned...)
			} else {
				out = append(out, replaceSlots(n.Children, fills)...)
			}
			continue
		}
		n.Children = replaceSlots(n.Children, fills)
		out = append(out, n)
	}
	return out
}

func (v View) expandUse(n Node, values map[string]string) ([]Node, map[string]string, error) {
	part, err := v.component(n.Attr("name", ""))
	if err != nil {
		return nil, nil, err
	}
	names, err := componentPropNames(part.Attr("props", ""))
	if err != nil {
		return nil, nil, err
	}
	declared := map[string]bool{}
	local := make(map[string]string, len(values)+len(names))
	for key, value := range values {
		if !strings.HasPrefix(key, "props.") {
			local[key] = value
		}
	}
	for _, name := range names {
		declared[name] = true
		raw, ok := n.Attrs[name]
		if !ok {
			return nil, nil, fmt.Errorf("component %q missing prop %q", part.Attr("name", ""), name)
		}
		local["props."+name] = interpolate(raw, values)
	}
	for key := range n.Attrs {
		if key != "name" && !declared[key] {
			return nil, nil, fmt.Errorf("undeclared component prop %q", key)
		}
	}
	declaredSlots := map[string]bool{}
	var findSlots func([]Node)
	findSlots = func(nodes []Node) {
		for _, child := range nodes {
			if child.Kind == "slot" {
				declaredSlots[child.Attr("name", "")] = true
			}
			findSlots(child.Children)
		}
	}
	findSlots(part.Children)
	fills := map[string][]Node{}
	for _, child := range n.Children {
		slot := child.Attr("slot", "")
		if !declaredSlots[slot] {
			return nil, nil, fmt.Errorf("unknown component slot %q", slot)
		}
		fills[slot] = child.Children
	}
	return replaceSlots(part.Children, fills), local, nil
}
