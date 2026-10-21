// Package sml parses and renders declarative terminal views.
// SML has no embedded code execution.
package sml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxFileSize = 64 * 1024
const maxDepth = 16
const maxNodes = 800

var validName = regexp.MustCompile("^[A-Za-z0-9_-]{1,48}$")
var binding = regexp.MustCompile("\\{\\{\\s*([A-Za-z_][A-Za-z0-9_.]*)\\s*\\}\\}")
var allowed = map[string]bool{
	"screen": true, "component": true, "column": true, "row": true, "grid": true,
	"text": true, "banner": true, "stat": true, "metric": true, "divider": true,
	"spacer": true, "panel": true, "badge": true, "progress": true, "include": true,
	"if": true, "when": true, "repeat": true, "table": true, "prompt": true, "alert": true,
	"gradient": true, "button": true, "link": true, "image": true, "gif": true,
	"gauge": true, "sparkline": true, "plugin-action": true,
	"toggle": true, "choice": true, "input": true, "number": true,
	"section": true, "field": true, "empty": true, "derive": true, "otherwise": true,
}

// Node is a checked presentation element, not an executable instruction.
type Node struct {
	Kind     string
	Attrs    map[string]string
	Text     string
	Children []Node
}

func (n Node) Attr(key, fallback string) string {
	if v, ok := n.Attrs[key]; ok {
		return v
	}
	return fallback
}

type View struct {
	Name   string
	Nodes  []Node
	assets string
}

func (v View) WithAssets(dir string) View { v.assets = dir; return v }

// Parse rejects invalid XML and unknown components before publishing a view.
func Parse(data []byte) (View, error) {
	if len(data) > maxFileSize || !utf8.Valid(data) {
		return View{}, errors.New("invalid SML size or UTF-8")
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	var root *Node
	stack := []*Node{}
	count := 0
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return View{}, fmt.Errorf("invalid SML XML: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			count++
			if count > maxNodes || len(stack) >= maxDepth {
				return View{}, errors.New("SML node/depth limit exceeded")
			}
			if !allowed[t.Name.Local] || t.Name.Space != "" {
				return View{}, fmt.Errorf("unsupported SML component <%s>", t.Name.Local)
			}
			n := &Node{Kind: t.Name.Local, Attrs: map[string]string{}}
			for _, a := range t.Attr {
				if a.Name.Space != "" {
					return View{}, errors.New("SML namespaces are not supported")
				}
				if _, exists := n.Attrs[a.Name.Local]; exists {
					return View{}, errors.New("duplicate SML attribute")
				}
				n.Attrs[a.Name.Local] = a.Value
			}
			if len(stack) == 0 {
				if root != nil {
					return View{}, errors.New("SML requires exactly one root")
				}
				root = n
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, *n)
				n = &parent.Children[len(parent.Children)-1]
			}
			stack = append(stack, n)
		case xml.CharData:
			if len(stack) == 0 {
				if strings.TrimSpace(string(t)) != "" {
					return View{}, errors.New("unexpected text outside SML root")
				}
			} else {
				stack[len(stack)-1].Text += string(t)
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return View{}, errors.New("unbalanced SML XML")
			}
			stack = stack[:len(stack)-1]
		case xml.ProcInst, xml.Directive:
			return View{}, errors.New("SML directives and processing instructions are not supported")
		}
	}
	if root == nil || root.Kind != "screen" {
		return View{}, errors.New("SML root must be <screen name=\"...\">")
	}
	version := root.Attr("version", "")
	if version != "" && version != "1" && version != "2" && version != "3" {
		return View{}, errors.New("unsupported SML version")
	}
	name := root.Attr("name", "")
	if !validName.MatchString(name) {
		return View{}, errors.New("invalid SML view name")
	}
	if len(stack) != 0 {
		return View{}, errors.New("incomplete SML XML")
	}
	view := View{Name: name, Nodes: root.Children}
	if version != "3" && requiresSML3(view.Nodes) {
		return View{}, errors.New("reactive SML components require version 3")
	}
	if err := validateUI(view.Nodes, name); err != nil {
		return View{}, err
	}
	if err := validateStyleSystem(view.Nodes); err != nil {
		return View{}, err
	}
	if name == "appearance" {
		if _, err := view.GradientStyles(); err != nil {
			return View{}, err
		}
	} else {
		for _, node := range view.Nodes {
			if node.Kind == "gradient" {
				return View{}, errors.New("gradient metadata belongs in appearance.sml")
			}
		}
	}
	return view, nil
}
func safe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || (r >= 32 && r != 127 && (r < 128 || r > 159)) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
func interpolate(s string, values map[string]string) string {
	s = binding.ReplaceAllStringFunc(s, func(token string) string {
		name := binding.FindStringSubmatch(token)[1]
		v := values[name]
		if len(v) > 4096 {
			v = v[:4096]
		}
		return safe(v)
	})
	return safe(s)
}
func boundedInt(value string, fallback, lo, hi int) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
func trim(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}
func pad(s string, w int) string {
	n := utf8.RuneCountInString(s)
	if n >= w {
		return trim(s, w)
	}
	return s + strings.Repeat(" ", w-n)
}
func (v View) component(name string) (Node, error) {
	if !validName.MatchString(name) {
		return Node{}, errors.New("invalid SML include name")
	}
	if v.assets == "" {
		return Node{}, errors.New("SML components need an assets folder")
	}
	p := filepath.Join(v.assets, "components", name+".sml")
	info, err := os.Lstat(p)
	if err != nil {
		return Node{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxFileSize {
		return Node{}, errors.New("invalid SML component file")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return Node{}, err
	}
	d := xml.NewDecoder(bytes.NewReader(b))
	// Component documents are parsed through the same strict view parser,
	// then validated as a component root without scripting support.
	var root struct {
		XMLName xml.Name
		Name    string
		Inner   string
	}
	_ = root
	var wrapper []byte
	wrapper = append(wrapper, []byte("<screen name=\"component\" version=\"3\">")...)
	wrapper = append(wrapper, b...)
	wrapper = append(wrapper, []byte("</screen>")...)
	parsed, err := Parse(wrapper)
	if err != nil {
		return Node{}, err
	}
	if len(parsed.Nodes) != 1 || parsed.Nodes[0].Kind != "component" || parsed.Nodes[0].Attr("name", "") != name {
		return Node{}, errors.New("SML component name/file mismatch")
	}
	_ = d
	return parsed.Nodes[0], nil
}
func (v View) lines(nodes []Node, values map[string]string, width, depth int) ([]string, error) {
	if depth > maxDepth {
		return nil, errors.New("SML include recursion/depth exceeded")
	}
	out := []string{}
	for _, node := range nodes {
		lines, err := v.node(node, values, width, depth+1)
		if err != nil {
			return nil, err
		}
		out = append(out, lines...)
		if len(out) > 3000 {
			return nil, errors.New("SML output row limit exceeded")
		}
	}
	return out, nil
}
func (v View) node(n Node, values map[string]string, width, depth int) ([]string, error) {
	if depth > maxDepth {
		return nil, errors.New("SML nesting limit exceeded")
	}
	width = max(12, min(width, 240))
	switch n.Kind {
	case "gradient", "derive":
		return nil, nil
	case "button", "link", "plugin-action":
		return []string{renderActionControl(n, values, width)}, nil
	case "toggle", "choice":
		return []string{renderSessionControl(n, values, width)}, nil
	case "input", "number":
		return []string{renderSessionInput(n, values, width)}, nil
	case "image", "gif":
		return v.renderTerminalMedia(n, width)
	case "gauge":
		return renderGauge(n, values, width), nil
	case "sparkline":
		return renderSparkline(n, values, width), nil
	case "screen", "component", "column":
		return v.lines(n.Children, values, width, depth)
	case "section":
		title := trim(interpolate(n.Attr("title", ""), values), width)
		out := []string{}
		if title != "" {
			out = append(out, styleMetric(n, title, width), styleDivider(n, min(width, 60)))
		}
		children, err := v.lines(n.Children, values, width, depth)
		if err != nil {
			return nil, err
		}
		return append(out, children...), nil
	case "field":
		label := interpolate(n.Attr("label", ""), values)
		name := n.Attr("key", "")
		value := values[name]
		if value == "" {
			value = interpolate(n.Attr("fallback", "—"), values)
		}
		suffix := interpolate(n.Attr("unit", ""), values)
		if suffix != "" {
			value += " " + suffix
		}
		return []string{styleMetric(n, label+": "+safe(value), width)}, nil
	case "empty":
		name := n.Attr("field", "")
		if values[name] != "" {
			return nil, nil
		}
		return v.lines(n.Children, values, width, depth)
	case "text", "badge", "alert", "prompt":
		content := interpolate(n.Attr("value", strings.TrimSpace(n.Text)), values)
		if n.Kind == "prompt" {
			content = interpolate(n.Attr("prefix", "> "), values)
		}
		if n.Kind == "badge" || n.Kind == "alert" {
			return []string{styleBadge(n, content, width)}, nil
		}
		return []string{styleMetric(n, content, width)}, nil
	case "stat":
		return []string{trim(interpolate(n.Text, values)+": "+values[n.Attr("value", "")], width)}, nil
	case "metric":
		return []string{styleMetric(n, interpolate(n.Attr("label", ""), values)+": "+interpolate(n.Attr("value", ""), values), width)}, nil
	case "divider":
		return []string{styleDivider(n, width)}, nil
	case "spacer":
		return make([]string, boundedInt(n.Attr("size", "1"), 1, 0, 5)), nil
	case "if", "when":
		field := n.Attr("field", n.Attr("value", ""))
		value := values[field]
		matched := false
		if n.Kind == "if" {
			matched = value == n.Attr("equals", "true")
		} else {
			matched = value != "" && value != "0" && value != "false" && value != "no"
		}
		primary := n.Children
		var alternative []Node
		for index, child := range n.Children {
			if child.Kind == "otherwise" {
				primary = n.Children[:index]
				alternative = child.Children
				break
			}
		}
		if !matched {
			return v.lines(alternative, values, width, depth)
		}
		return v.lines(primary, values, width, depth)
	case "otherwise":
		return nil, errors.New("SML otherwise must be nested in a condition")
	case "include":
		part, err := v.component(n.Attr("name", ""))
		if err != nil {
			return nil, err
		}
		return v.node(part, values, width, depth+1)
	case "row", "grid":
		if len(n.Children) == 0 {
			return nil, nil
		}
		columns := len(n.Children)
		if n.Kind == "grid" {
			columns = boundedInt(n.Attr("columns", "3"), 3, 1, 8)
		}
		if width < boundedInt(n.Attr("collapse-below", "65"), 65, 20, 240) {
			columns = 1
		}
		gap := boundedInt(n.Attr("gap", "2"), 2, 0, 8)
		columns = min(columns, len(n.Children))
		if (width-gap*(columns-1))/columns < 14 {
			columns = 1
		}
		subwidth := max(12, (width-gap*(columns-1))/columns)
		out := []string{}
		for i := 0; i < len(n.Children); i += columns {
			count := min(columns, len(n.Children)-i)
			groups := make([][]string, count)
			maxrows := 0
			for j := 0; j < count; j++ {
				ls, err := v.node(n.Children[i+j], values, subwidth, depth+1)
				if err != nil {
					return nil, err
				}
				groups[j] = ls
				if len(ls) > maxrows {
					maxrows = len(ls)
				}
			}
			for row := 0; row < maxrows; row++ {
				parts := []string{}
				for j, g := range groups {
					value := ""
					if row < len(g) {
						value = g[row]
					}
					if j < count-1 {
						value = pad(value, subwidth)
					}
					parts = append(parts, value)
				}
				out = append(out, strings.TrimRight(strings.Join(parts, strings.Repeat(" ", gap)), " "))
			}
		}
		return out, nil
	case "panel":
		padding := boundedInt(n.Attr("padding", "1"), 1, 0, 3)
		inner := width - 2 - 2*padding
		if n.Attr("border", "round") == "none" {
			inner = width - padding
		}
		content, err := v.lines(n.Children, values, max(12, inner), depth)
		if err != nil {
			return nil, err
		}
		return stylePanel(n, content, interpolate(n.Attr("title", ""), values), width), nil
	case "progress":
		pct := boundedInt(interpolate(n.Attr("value", "0"), values), 0, 0, 100)
		barwidth := min(24, max(8, width-20))
		filled := pct * barwidth / 100
		return []string{trim(interpolate(n.Attr("label", "Progress"), values)+" ["+strings.Repeat("━", filled)+strings.Repeat("─", barwidth-filled)+"] "+strconv.Itoa(pct)+"%", width)}, nil
	case "banner":
		name := n.Attr("src", "logo.txt")
		if !validName.MatchString(strings.TrimSuffix(name, ".txt")) {
			return nil, errors.New("invalid SML banner asset name")
		}
		if v.assets == "" {
			return nil, errors.New("SML banner needs an assets folder")
		}
		path := filepath.Join(v.assets, "art", name)
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > maxFileSize {
			return nil, errors.New("invalid SML banner file")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		// Center the banner as one block. Per-line centering shears ASCII art when
		// lines have different lengths (classic TUI "broken logo" look).
		raw := []string{}
		maxW := 0
		for _, line := range strings.Split(strings.TrimRight(safe(string(b)), "\r\n"), "\n") {
			line = trim(line, width)
			raw = append(raw, line)
			if w := utf8.RuneCountInString(line); w > maxW {
				maxW = w
			}
		}
		padLeft := 0
		if n.Attr("align", "") == "center" {
			padLeft = max(0, (width-maxW)/2)
		}
		out := make([]string, len(raw))
		for i, line := range raw {
			if padLeft > 0 {
				line = strings.Repeat(" ", padLeft) + line
			}
			out[i] = line
		}
		return out, nil
	case "repeat":
		key := n.Attr("source", "")
		alias := n.Attr("as", "item")
		if !validName.MatchString(alias) {
			return nil, errors.New("invalid SML repeat alias")
		}
		items := strings.Split(values[key], ",")
		if values[key] == "" {
			return nil, nil
		}
		limit := min(len(items), boundedInt(n.Attr("limit", "30"), 30, 0, 100))
		out := []string{}
		for i := 0; i < limit; i++ {
			local := make(map[string]string, len(values)+1)
			for k, val := range values {
				local[k] = val
			}
			local[alias] = strings.TrimSpace(items[i])
			lines, err := v.lines(n.Children, local, width, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, lines...)
		}
		return out, nil
	case "table":
		return renderDataTable(n, values, width)
	}
	return nil, fmt.Errorf("unsupported SML element <%s>", n.Kind)
}
func (v View) Render(values map[string]string) []string { return v.RenderWidth(values, 80) }

// RenderChecked returns rendered lines and any render-time error.
// Unlike RenderWidth it does not convert errors to visible error text.
func (v View) RenderChecked(values map[string]string, width int) ([]string, error) {
	return v.lines(v.Nodes, v.ResolveBindings(values), width, 0)
}
func (v View) RenderWidth(values map[string]string, width int) []string {
	out, err := v.RenderChecked(values, width)
	if err != nil {
		return []string{"SML render error: " + safe(err.Error())}
	}
	return out
}
func LoadDir(dir string) (map[string]View, error) {
	directory := filepath.Join(dir, "views")
	items, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]View{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(items) > 100 {
		return nil, errors.New("too many SML assets")
	}
	views := map[string]View{}
	for _, entry := range items {
		if !strings.HasSuffix(entry.Name(), ".sml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".sml")
		if !validName.MatchString(name) {
			return nil, fmt.Errorf("invalid SML filename: %s", entry.Name())
		}
		path := filepath.Join(directory, entry.Name())
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > maxFileSize {
			return nil, fmt.Errorf("invalid SML asset: %s", entry.Name())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		view, err := Parse(content)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		if view.Name != name {
			return nil, fmt.Errorf("SML view name/file mismatch: %s", name)
		}
		view.assets = dir
		// Compile and resolve includes before publishing a new immutable snapshot.
		if err = validateIncludes(view, view.Nodes, 0, map[string]bool{}); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		views[name] = view
	}
	return views, nil
}
func validateIncludes(v View, nodes []Node, depth int, seen map[string]bool) error {
	if depth > maxDepth {
		return errors.New("SML include depth exceeded")
	}
	for _, node := range nodes {
		if node.Kind == "include" {
			name := node.Attr("name", "")
			if seen[name] {
				return fmt.Errorf("SML include cycle: %s", name)
			}
			part, err := v.component(name)
			if err != nil {
				return err
			}
			seen[name] = true
			if err = validateIncludes(v, part.Children, depth+1, seen); err != nil {
				return err
			}
			delete(seen, name)
		}
		if err := validateIncludes(v, node.Children, depth+1, seen); err != nil {
			return err
		}
	}
	return nil
}
