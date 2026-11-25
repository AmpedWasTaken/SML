package engine

import (
	"fmt"
	"strconv"
)

// boxViewport clips a rendered box in display rows. Scroll state belongs
// to the host, and is never modified by markup.
func boxViewport(n Node, values map[string]string, total int) (start, end int) {
	end = total
	mode := n.Attr("overflow", "visible")
	if mode == "visible" {
		return
	}
	limit := boundedInt(n.Attr("max-height", "3000"), 3000, 1, 3000)
	if mode == "auto" {
		offset, err := strconv.Atoi(values[n.Attr("scroll-key", "")])
		if err == nil {
			start = max(0, min(offset, max(0, total-limit)))
		}
	}
	end = min(total, start+limit)
	return
}

func clipBoxViewport(n Node, values map[string]string, block layoutBlock) layoutBlock {
	start, end := boxViewport(n, values, len(block.lines))
	if start == 0 && end == len(block.lines) {
		return block
	}
	out := layoutBlock{lines: block.lines[start:end]}
	for _, spot := range block.spots {
		if spot.Line >= start && spot.Line < end {
			spot.Line -= start
			out.spots = append(out.spots, spot)
		}
	}
	for _, element := range block.elements {
		r := element.Bounds
		top, bottom := max(r.Y, start), min(r.Y+r.Height, end)
		if bottom > top {
			element.Bounds.Y = top - start
			element.Bounds.Height = bottom - top
			out.elements = append(out.elements, element)
		}
	}
	return out
}

func validateBoxScrollKey(value string) error {
	if !sessionKey.MatchString(value) {
		return fmt.Errorf("box scroll-key must name a state.* binding")
	}
	return nil
}
