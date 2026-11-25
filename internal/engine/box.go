package engine

import (
	"fmt"
	"strconv"
	"strings"
)

func validateBox(n Node) error {
	if n.Kind != "box" {
		return nil
	}
	for key, value := range n.Attrs {
		switch key {
		case "width":
			width, err := strconv.Atoi(value)
			if err != nil || width < 12 || width > 240 {
				return fmt.Errorf("SML box width must be 12..240")
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
	return nil
}

// The box is a bounded responsive container, not a fixed terminal size.
func boxDimensions(n Node, available int) (width, left int) {
	width = min(available, boundedInt(n.Attr("width", ""), available, 12, 240))
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
