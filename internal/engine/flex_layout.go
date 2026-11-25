package engine

import (
	"errors"
	"fmt"
)

// flexWidths allocates bounded terminal cells. Shrinking is opt-in, so
// existing flex rows retain their original vertical fallback behavior.
func flexWidths(children []Node, width, gap int) ([]int, int, bool) {
	count := len(children)
	if count < 1 || count > 8 {
		return nil, 0, false
	}
	available := width - gap*(count-1)
	if available < 14*count {
		return nil, 0, false
	}
	widths := make([]int, count)
	grows := make([]int, count)
	floors := make([]int, count)
	shrinks := make([]int, count)
	total, grow := 0, 0
	for i, child := range children {
		widths[i] = max(14, resolveBoxLength(child.Attr("flex-basis", ""), available, 14))
		floors[i] = max(14, resolveBoxLength(child.Attr("min-width", ""), available, 14))
		widths[i] = max(widths[i], floors[i])
		shrinks[i] = boundedInt(child.Attr("flex-shrink", "0"), 0, 0, 16) * widths[i]
		grows[i] = boundedInt(child.Attr("flex-grow", "1"), 1, 0, 16)
		total += widths[i]
		grow += grows[i]
	}
	if total > available {
		if !shrinkFlexWidths(widths, floors, shrinks, total-available) {
			return nil, 0, false
		}
		return widths, 0, true
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

// Distribute the deficit in proportion to shrink factor times base size,
// freezing items at their floor and redistributing the remaining deficit.
func shrinkFlexWidths(widths, floors, weights []int, deficit int) bool {
	for deficit > 0 {
		total := 0
		for i, weight := range weights {
			if widths[i] > floors[i] {
				total += weight
			}
		}
		if total == 0 {
			return false
		}
		initial := deficit
		for i, weight := range weights {
			if weight == 0 || widths[i] <= floors[i] {
				continue
			}
			delta := min(deficit, min(widths[i]-floors[i], initial*weight/total))
			widths[i] -= delta
			deficit -= delta
		}
		if deficit == initial {
			// Integer-cell rounding: always make progress.
			for i, weight := range weights {
				if weight > 0 && widths[i] > floors[i] {
					widths[i]--
					deficit--
					break
				}
			}
		}
	}
	return true
}

// Wrapping considers the natural bases before shrinking, like CSS flex
// line formation. A line holds at most eight panes.
func flexNaturalFits(children []Node, width, gap int) bool {
	if len(children) > 8 || len(children) == 0 {
		return false
	}
	available := width - gap*(len(children)-1)
	if available < 14*len(children) {
		return false
	}
	total := 0
	for _, child := range children {
		base := resolveBoxLength(child.Attr("flex-basis", ""), available, 14)
		floor := resolveBoxLength(child.Attr("min-width", ""), available, 14)
		total += max(14, max(base, floor))
	}
	return total <= available
}

func (v View) renderFlex(n Node, values map[string]string, width, depth int, path string) (layoutBlock, error) {
	result := layoutBlock{}
	gap := boundedInt(n.Attr("gap", "2"), 2, 0, 8)
	rowGap := boundedInt(n.Attr("row-gap", "1"), 1, 0, 8)
	children := make([]Node, len(n.Children))
	for i, child := range n.Children {
		children[i] = v.responsiveNode(child)
	}
	if len(children) == 0 {
		return result, nil
	}
	stack := n.Attr("flex-direction", "row") == "column" || width < boundedInt(n.Attr("collapse-below", "20"), 20, 20, 240)
	wrap := n.Attr("flex-wrap", "nowrap") == "wrap"
	if _, _, fits := flexWidths(children, width, gap); !wrap && !fits {
		stack = true
	}
	for start := 0; start < len(children); {
		end := len(children)
		if stack {
			end = start + 1
		} else if wrap {
			end = start + 1
			for end < len(children) && flexNaturalFits(children[start:end+1], width, gap) {
				end++
			}
		}
		line := children[start:end]
		widths, free, fits := flexWidths(line, width, gap)
		if stack || !fits {
			// A single oversized item still gets the full available width.
			// For a wrapped line, retry with one item to preserve geometry.
			end = start + 1
			line = children[start:end]
			widths, free = []int{width}, 0
		}
		part, err := v.renderFlexLine(n, line, widths, free, gap, width, depth, start, path, values)
		if err != nil {
			return layoutBlock{}, err
		}
		if start > 0 {
			result.lines = append(result.lines, make([]string, rowGap)...)
		}
		offsetBlock(&result, part, 0, len(result.lines))
		result.lines = append(result.lines, part.lines...)
		if len(result.lines) > 3000 {
			return layoutBlock{}, errors.New("SML flex layout exceeds output limit")
		}
		start = end
	}
	return result, nil
}

func (v View) renderFlexLine(n Node, children []Node, widths []int, free, gap, width, depth, start int, path string, values map[string]string) (layoutBlock, error) {
	result := layoutBlock{}
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
		part, err := v.renderLayoutNode(child, values, widths[i], depth+1, fmt.Sprintf("%s/%d:%s", path, start+i, child.Kind))
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
