package engine

import (
	"fmt"
	"strings"
)

func validateBox(n Node) error {
	if n.Kind != "box" {
		return nil
	}
	for key, value := range n.Attrs {
		switch key {
		case "width", "min-width", "max-width", "flex-grow", "flex-basis", "flex-shrink", "max-height", "overflow", "position", "left", "top", "z-index":
			if err := validateLayoutStyleValue(key, value); err != nil {
				return err
			}
		case "scroll-key":
			if err := validateBoxScrollKey(value); err != nil {
				return err
			}
		case "align":
			if err := validateStyleAttribute(key, value); err != nil {
				return err
			}
		case "foreground", "background":
			if err := validateStyleAttribute(key, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported SML box attribute %q", key)
		}
	}
	for _, key := range []string{"left", "top", "z-index"} {
		if _, ok := n.Attrs[key]; ok && n.Attr("position", "static") != "absolute" {
			return fmt.Errorf("%s requires position=absolute", key)
		}
	}
	return nil
}

// The box is a bounded responsive container, not a fixed terminal size.
func boxDimensions(n Node, available int) (width, left int) {
	preferred := resolveBoxLength(n.Attr("width", ""), available, available)
	low := resolveBoxLength(n.Attr("min-width", ""), available, 12)
	high := resolveBoxLength(n.Attr("max-width", ""), available, available)
	width = min(available, max(12, min(max(preferred, low), max(low, high))))
	switch n.Attr("align", "left") {
	case "center":
		left = (available - width) / 2
	case "right":
		left = available - width
	}
	return width, left
}

func boxLines(lines []string, width, left int) []string {
	out := make([]string, len(lines))
	prefix := strings.Repeat(" ", left)
	for i, line := range lines {
		out[i] = prefix + PadCells(line, width)
	}
	return out
}
