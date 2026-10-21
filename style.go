package sml

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// SML v3 styling is defined directly on visual elements. No CSS-like
// declarations, external theme imports, or executable style expressions.
var styleAttributes = map[string]bool{
	"border": true, "padding": true, "align": true, "case": true,
	"variant": true, "tone": true,
}
var styleTargets = map[string]bool{
	"text": true, "badge": true, "alert": true, "metric": true,
	"field": true, "section": true, "panel": true, "divider": true,
	"button": true, "link": true, "plugin-action": true,
}

// Each visual element accepts only attributes used by its actual renderer.
// Reject mistyped/unsupported properties rather than silently ignoring them.
var inlineProps = map[string]map[string]bool{
	"text":          {"value": true, "align": true, "case": true, "padding": true, "tone": true},
	"badge":         {"value": true, "align": true, "case": true, "padding": true, "tone": true, "variant": true},
	"alert":         {"value": true, "align": true, "case": true, "padding": true, "tone": true, "variant": true},
	"metric":        {"label": true, "value": true, "align": true, "case": true, "padding": true, "tone": true},
	"field":         {"key": true, "label": true, "unit": true, "fallback": true, "align": true, "case": true, "padding": true, "tone": true},
	"section":       {"title": true, "align": true, "case": true, "padding": true, "tone": true, "variant": true},
	"panel":         {"title": true, "border": true, "align": true, "case": true, "padding": true, "tone": true},
	"divider":       {"variant": true},
	"button":        {"label": true, "page": true, "command": true, "enabled-when": true, "align": true, "case": true, "padding": true, "variant": true},
	"link":          {"label": true, "page": true, "command": true, "enabled-when": true, "align": true, "case": true, "padding": true, "variant": true},
	"plugin-action": {"label": true, "plugin": true, "action": true, "arg": true, "source": true, "bindings": true, "enabled-when": true, "align": true, "case": true, "padding": true, "variant": true},
}

func validateStyleAttribute(key, value string) error {
	switch key {
	case "border":
		switch value {
		case "round", "single", "double", "heavy", "ascii", "none":
			return nil
		}
	case "padding":
		switch value {
		case "0", "1", "2", "3":
			return nil
		}
	case "align":
		switch value {
		case "left", "center", "right":
			return nil
		}
	case "case":
		switch value {
		case "none", "upper", "lower":
			return nil
		}
	case "variant":
		switch value {
		case "default", "outline", "solid", "minimal", "muted", "dashed", "dotted", "double", "heavy", "ascii":
			return nil
		}
	case "tone":
		switch value {
		case "neutral", "info", "success", "warning", "danger":
			return nil
		}
	}
	return fmt.Errorf("invalid SML style %s=%q", key, value)
}

// validateStyleSystem rejects unknown styling values at parse time. Existing
// "align" on nonstyled elements is maintained for legacy SML compatibility.
// Direct styling never changes authorization or executes code.
func validateStyleSystem(nodes []Node) error {
	var walk func([]Node) error
	walk = func(parts []Node) error {
		for _, node := range parts {
			if _, styleTag := node.Attrs["style"]; styleTag && node.Kind != "table" {
				return fmt.Errorf("SML <%s> uses unsupported style=; set attributes directly", node.Kind)
			}
			if allowed, ok := inlineProps[node.Kind]; ok {
				for key := range node.Attrs {
					if !allowed[key] {
						return fmt.Errorf("unsupported attribute %q on SML <%s>", key, node.Kind)
					}
				}
			}
			for key := range styleAttributes {
				value, ok := node.Attrs[key]
				if !ok {
					continue
				}
				if !styleTargets[node.Kind] {
					if key == "align" {
						continue
					}
					return fmt.Errorf("styling not supported on <%s>", node.Kind)
				}
				if err := validateStyleAttribute(key, value); err != nil {
					return err
				}
			}
			if err := walk(node.Children); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(nodes)
}

func styleCase(s, kind string) string {
	switch kind {
	case "upper":
		return strings.ToUpper(s)
	case "lower":
		return strings.ToLower(s)
	default:
		return s
	}
}
func styleAlign(s, align string, width int) string {
	s = trim(s, width)
	slack := width - utf8.RuneCountInString(s)
	if slack <= 0 {
		return s
	}
	switch align {
	case "center":
		return strings.Repeat(" ", slack/2) + s
	case "right":
		return strings.Repeat(" ", slack) + s
	default:
		return s
	}
}
func stylePad(n Node, s string, width int) string {
	padding := boundedInt(n.Attr("padding", "0"), 0, 0, 3)
	usable := max(0, width-padding*2)
	value := styleAlign(styleCase(s, n.Attr("case", "none")), n.Attr("align", "left"), usable)
	return strings.Repeat(" ", padding) + value
}
func styleTone(n Node) string {
	switch n.Attr("tone", "neutral") {
	case "info":
		return "[i] "
	case "success":
		return "[OK] "
	case "warning":
		return "[!] "
	case "danger":
		return "[X] "
	default:
		return ""
	}
}
func styleBadge(n Node, content string, width int) string {
	content = styleCase(content, n.Attr("case", "none"))
	switch n.Attr("variant", "default") {
	case "outline":
		content = "[ " + content + " ]"
	case "solid":
		content = "<< " + content + " >>"
	case "minimal": // no chrome
	case "muted":
		content = "( " + content + " )"
	default:
		content = "● " + content
	}
	return trim(stylePad(n, styleTone(n)+content, width), width)
}
func styleMetric(n Node, content string, width int) string {
	return trim(stylePad(n, styleTone(n)+content, width), width)
}
func styleAction(n Node, label string, width int) string {
	label = styleCase(label, n.Attr("case", "none"))
	switch n.Attr("variant", "default") {
	case "solid":
		label = "<< " + label + " >>"
	case "minimal":
		label = label + " >"
	case "muted":
		label = "( " + label + " )"
	default:
		label = "[ " + label + " ]"
	}
	return trim(stylePad(n, label, width), width)
}
func styleDivider(n Node, width int) string {
	glyph := "─"
	switch n.Attr("variant", "default") {
	case "dashed":
		glyph = "┄"
	case "dotted":
		glyph = "·"
	case "double":
		glyph = "═"
	case "heavy":
		glyph = "━"
	case "ascii":
		glyph = "-"
	}
	return strings.Repeat(glyph, max(0, width))
}

type panelBorder struct {
	tl, tr, bl, br, h, v string
}

func styleBorder(name string) panelBorder {
	switch name {
	case "single":
		return panelBorder{"┌", "┐", "└", "┘", "─", "│"}
	case "double":
		return panelBorder{"╔", "╗", "╚", "╝", "═", "║"}
	case "heavy":
		return panelBorder{"┏", "┓", "┗", "┛", "━", "┃"}
	case "ascii":
		return panelBorder{"+", "+", "+", "+", "-", "|"}
	default:
		return panelBorder{"╭", "╮", "╰", "╯", "─", "│"}
	}
}
func stylePanel(n Node, content []string, title string, width int) []string {
	width = max(12, min(width, 240))
	padding := boundedInt(n.Attr("padding", "1"), 1, 0, 3)
	if n.Attr("border", "round") == "none" {
		out := make([]string, 0, len(content)+1)
		if title != "" {
			out = append(out, stylePad(n, styleTone(n)+title, width))
		}
		for _, line := range content {
			out = append(out, strings.Repeat(" ", padding)+trim(line, width-padding))
		}
		return out
	}
	border := styleBorder(n.Attr("border", "round"))
	inner := max(0, width-2-2*padding)
	header := ""
	if title != "" {
		header = " " + trim(styleCase(title, n.Attr("case", "none")), max(0, width-6)) + " "
	}
	if utf8.RuneCountInString(header) > width-2 {
		header = trim(header, width-2)
	}
	horizontal := header + strings.Repeat(border.h, max(0, width-2-utf8.RuneCountInString(header)))
	out := []string{border.tl + horizontal + border.tr}
	blank := border.v + strings.Repeat(" ", width-2) + border.v
	for i := 0; i < padding-1; i++ {
		out = append(out, blank)
	}
	for _, line := range content {
		line = trim(line, inner)
		if n.Attr("align", "left") != "left" {
			line = styleAlign(line, n.Attr("align", "left"), inner)
		}
		out = append(out, border.v+strings.Repeat(" ", padding)+pad(line, inner)+strings.Repeat(" ", padding)+border.v)
	}
	for i := 0; i < padding-1; i++ {
		out = append(out, blank)
	}
	out = append(out, border.bl+strings.Repeat(border.h, width-2)+border.br)
	return out
}

// New inline style attributes opt a view into SML 3, without a separate
// declaration. "align" predates v3 and remains valid in older views.
func hasStyleAttributes(n Node) bool {
	for key := range styleAttributes {
		if key == "align" {
			continue
		}
		if _, ok := n.Attrs[key]; ok {
			return true
		}
	}
	return false
}
