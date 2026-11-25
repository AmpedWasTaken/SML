package engine

import (
	"errors"
	"fmt"
)

// flexWidths allocates bounded cell widths. A row whose bases cannot fit
// stacks rather than clipping its interactive children.
func flexWidths(children []Node, width, gap int) ([]int, int, bool) {
	count := len(children)
	if count < 1 || count > 8 {
		return nil, 0, false
	}
	available := width - gap*(count-1)
	widths := make([]int, count)
	grows := make([]int, count)
	total, grow := 0, 0
	for i, child := range children {
		widths[i] = max(14, resolveBoxLength(child.Attr("flex-basis", ""), available, 14))
		grows[i] = boundedInt(child.Attr("flex-grow", "1"), 1, 0, 16)
		total += widths[i]
		grow += grows[i]
	}
	if total > available {
		return nil, 0, false
	}
	free := available - total
	for i := range widths {
		if grow == 0 {
			break
		}
		add := free * grows[i] / grow
		widths[i] += add
		free -= add
		grow -= grows[i]
	}
	return widths, free, true
}

func (v View) renderFlex(n Node, values map[string]string, width, depth int, path string) (layoutBlock, error) {
	result := layoutBlock{}
	gap := boundedInt(n.Attr("gap", "2"), 2, 0, 8)
	children := make([]Node, len(n.Children))
	for i, child := range n.Children {
		children[i] = v.responsiveNode(child)
	}
	widths, free, fits := flexWidths(children, width, gap)
	stack := !fits || n.Attr("flex-direction", "row") == "column" || width < boundedInt(n.Attr("collapse-below", "20"), 20, 20, 240)
	if stack {
		for i, child := range children {
			if i > 0 {
				result.lines = append(result.lines, make([]string, boundedInt(n.Attr("row-gap", "1"), 1, 0, 8))...)
			}
			part, err := v.renderLayoutNode(child, values, width, depth+1, fmt.Sprintf("%s/%d:%s", path, i, child.Kind))
			if err != nil {
				return layoutBlock{}, err
			}
			offsetBlock(&result, part, 0, len(result.lines))
			result.lines = append(result.lines, part.lines...)
			if len(result.lines) > 3000 {
				return layoutBlock{}, errors.New("SML flex layout exceeds output limit")
			}
		}
		return result, nil
	}
	x, height := 0, 0
	justify := n.Attr("justify-content", "start")
	switch justify {
	case "center":
		x = free / 2
	case "end":
		x = free
	}
	parts := make([]locatedBlock, 0, len(children))
	for i, child := range children {
		part, err := v.renderLayoutNode(child, values, widths[i], depth+1, fmt.Sprintf("%s/%d:%s", path, i, child.Kind))
		if err != nil {
			return layoutBlock{}, err
		}
		height = max(height, len(part.lines))
		parts = append(parts, locatedBlock{block: part, x: x})
		x += widths[i] + gap
		if justify == "space-between" && i < len(children)-1 {
			spacing := free / (len(children) - i - 1)
			x += spacing
			free -= spacing
		}
	}
	for i := range parts {
		switch n.Attr("align-items", "start") {
		case "center":
			parts[i].y = (height - len(parts[i].block.lines)) / 2
		case "end":
			parts[i].y = height - len(parts[i].block.lines)
		}
		offsetBlock(&result, parts[i].block, parts[i].x, parts[i].y)
	}
	result.lines = renderCanvas(width, height, parts)
	return result, nil
}
