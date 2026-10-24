package engine

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var binding = regexp.MustCompile("\\{\\{\\s*([A-Za-z_][A-Za-z0-9_.]*)\\s*\\}\\}")

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
            // Bound input in bytes without slicing through the middle of an
            // UTF-8 sequence. A partially sliced emoji used to display as
            // a replacement character in panels, titles and table cells.
            v = v[:4096]
            for !utf8.ValidString(v) {
                v = v[:len(v)-1]
            }
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
	return ClipCells(s, max)
}
func pad(s string, w int) string {
	return PadCells(s, w)
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
	case "use":
		nodes, local, err := v.expandUse(n, values)
		if err != nil {
			return nil, err
		}
		return v.lines(nodes, local, width, depth+1)
	case "slot", "fill":
		return nil, errors.New("unexpanded SML slot or fill")
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
			if w := CellWidth(line); w > maxW {
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
