package sml

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Options configures static, ANSI-free rendering.
type Options struct {
	// Width is the terminal's available width in columns (default: 80).
	Width int
}

// Render writes a static frame. No user input, shell commands, or network calls.
func Render(w io.Writer, root *Node, options Options) error {
	if root == nil || root.Tag != "sml" {
		return fmt.Errorf("expected a parsed <sml> root")
	}
	width := options.Width
	if width == 0 {
		width = 80
	}
	if width < 24 || width > 240 {
		return fmt.Errorf("render width must be between 24 and 240")
	}
	lines, err := renderChildren(root.Children, width)
	if err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

// RenderString is convenient for snapshots, unit tests, and embedded TUIs.
func RenderString(root *Node, options Options) (string, error) {
	var out strings.Builder
	if err := Render(&out, root, options); err != nil {
		return "", err
	}
	return out.String(), nil
}

func renderChildren(children []*Node, width int) ([]string, error) {
	var lines []string
	for _, child := range children {
		frame, err := renderNode(child, width)
		if err != nil {
			return nil, err
		}
		lines = append(lines, frame...)
	}
	return lines, nil
}

func renderNode(node *Node, width int) ([]string, error) {
	switch node.Tag {
	case "text":
		return wrap(clean(node.Text), width), nil
	case "metric":
		label := clean(node.Attrs["label"])
		value := clean(node.Attrs["value"])
		if label == "" {
			return nil, fmt.Errorf("metric requires a label attribute")
		}
		if value == "" {
			return nil, fmt.Errorf("metric requires a value attribute")
		}
		pad := width - utf8.RuneCountInString(label) - utf8.RuneCountInString(value)
		if pad < 2 {
			return wrap(label+"  "+value, width), nil
		}
		return []string{label + strings.Repeat(" ", pad) + value}, nil
	case "panel":
		if width < 10 {
			return nil, fmt.Errorf("panel needs at least 10 columns")
		}
		lines, err := renderChildren(node.Children, width-4)
		if err != nil {
			return nil, err
		}
		title := clean(node.Attrs["title"])
		header := strings.Repeat("─", width-2)
		if title != "" {
			segment := " " + title + " "
			if utf8.RuneCountInString(segment) > width-4 {
				segment = truncate(segment, width-4)
			}
			header = segment + strings.Repeat("─", width-2-utf8.RuneCountInString(segment))
		}
		out := []string{"┌" + header + "┐"}
		if len(lines) == 0 {
			lines = []string{""}
		}
		for _, line := range lines {
			out = append(out, "│ "+fit(line, width-4)+" │")
		}
		out = append(out, "└"+strings.Repeat("─", width-2)+"┘")
		return out, nil
	case "grid":
		columns := 2
		if value, ok := node.Attrs["columns"]; ok {
			parsed, err := strconv.Atoi(value)
			if err != nil || parsed < 1 || parsed > 4 {
				return nil, fmt.Errorf("grid columns must be 1..4")
			}
			columns = parsed
		}
		for columns > 1 && (width-2*(columns-1))/columns < 12 {
			columns--
		}
		gap := 2
		cellWidth := (width - gap*(columns-1)) / columns
		var all []string
		for start := 0; start < len(node.Children); start += columns {
			end := start + columns
			if end > len(node.Children) {
				end = len(node.Children)
			}
			frames := make([][]string, 0, end-start)
			maxLines := 0
			for _, child := range node.Children[start:end] {
				frame, err := renderNode(child, cellWidth)
				if err != nil {
					return nil, err
				}
				frames = append(frames, frame)
				if len(frame) > maxLines {
					maxLines = len(frame)
				}
			}
			for i := 0; i < maxLines; i++ {
				parts := make([]string, len(frames))
				for j, frame := range frames {
					line := ""
					if i < len(frame) {
						line = frame[i]
					}
					parts[j] = fit(line, cellWidth)
				}
				all = append(all, strings.Join(parts, strings.Repeat(" ", gap)))
			}
		}
		return all, nil
	default:
		return nil, fmt.Errorf("cannot render element %q", node.Tag)
	}
}

func clean(raw string) string {
	var out strings.Builder
	for _, r := range raw {
		if unicode.IsControl(r) {
			if r == '\t' || r == '\n' || r == '\r' {
				out.WriteRune(' ')
			}
			continue
		}
		out.WriteRune(r)
	}
	return strings.Join(strings.Fields(out.String()), " ")
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

func fit(s string, width int) string {
	s = truncate(s, width)
	return s + strings.Repeat(" ", width-utf8.RuneCountInString(s))
}

func wrap(s string, width int) []string {
	if s == "" {
		return []string{""}
	}
	words := strings.Fields(s)
	lines := []string{}
	current := ""
	for _, word := range words {
		// Split long unbroken words instead of exceeding the terminal width.
		for utf8.RuneCountInString(word) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, truncate(word, width))
			word = string([]rune(word)[width:])
		}
		if word == "" {
			continue
		}
		if current == "" {
			current = word
		} else if utf8.RuneCountInString(current)+1+utf8.RuneCountInString(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
