package engine

import (
	"errors"
	"fmt"
	"sort"
)

type positionedLayer struct {
	block layoutBlock
	x, y  int
	width int
	z     int
}

// A relative box establishes a local positioning context. Absolute boxes
// are opaque layers, painted in ascending z-index and source order.
func (v View) renderPositionedBox(n Node, values map[string]string, width, depth int, path string) (layoutBlock, error) {
	base := layoutBlock{}
	layers := make([]positionedLayer, 0)
	for i, raw := range n.Children {
		child := v.responsiveNode(raw)
		childPath := fmt.Sprintf("%s/%d:%s", path, i, child.Kind)
		if child.Kind != "box" || child.Attr("position", "static") != "absolute" {
			part, err := v.renderLayoutNode(child, values, width, depth+1, childPath)
			if err != nil {
				return layoutBlock{}, err
			}
			offsetBlock(&base, part, 0, len(base.lines))
			base.lines = append(base.lines, part.lines...)
			if len(base.lines) > 3000 {
				return layoutBlock{}, errors.New("SML positioned layout exceeds row limit")
			}
			continue
		}
		if len(layers) >= 16 {
			return layoutBlock{}, errors.New("SML positioned box supports at most 16 layers")
		}
		x := min(max(0, width-12), boundedInt(child.Attr("left", "0"), 0, 0, 240))
		y := boundedInt(child.Attr("top", "0"), 0, 0, 3000)
		available := width - x
		// Absolute placement determines the origin, not the box's alignment.
		attrs := make(map[string]string, len(child.Attrs))
		for key, value := range child.Attrs {
			attrs[key] = value
		}
		attrs["align"] = "left"
		child.Attrs = attrs
		layerWidth, _ := boxDimensions(child, available)
		part, err := v.renderLayoutNode(child, values, available, depth+1, childPath)
		if err != nil {
			return layoutBlock{}, err
		}
		if len(part.lines) == 0 {
			continue
		}
		if y+len(part.lines) > 3000 {
			return layoutBlock{}, errors.New("SML positioned layer exceeds row limit")
		}
		layers = append(layers, positionedLayer{
			block: part, x: x, y: y, width: layerWidth,
			z: boundedInt(child.Attr("z-index", "0"), 0, 0, 100),
		})
	}
	sort.SliceStable(layers, func(i, j int) bool { return layers[i].z < layers[j].z })
	for _, layer := range layers {
		base = compositeLayer(base, layer, width)
	}
	return base, nil
}

// Erase full grapheme clusters crossing an overlay edge. Return the
// actual overwritten span so partially occluded controls are disabled.
func overwriteLayerCells(row []Cell, left int, overlay []Cell) (int, int) {
	right := left + len(overlay)
	start, end := left, right
	for start > 0 && start < len(row) && row[start].Continuation {
		start--
	}
	for end < len(row) && row[end].Continuation {
		end++
	}
	clear(row[start:end])
	copy(row[left:right], overlay)
	return start, end
}

func compositeLayer(base layoutBlock, layer positionedLayer, width int) layoutBlock {
	height := max(len(base.lines), layer.y+len(layer.block.lines))
	out := layoutBlock{lines: make([]string, height)}
	covered := make([]Rect, 0, len(layer.block.lines))
	for y := 0; y < height; y++ {
		line := ""
		if y < len(base.lines) {
			line = base.lines[y]
		}
		if y < layer.y || y >= layer.y+len(layer.block.lines) {
			out.lines[y] = line
			continue
		}
		row := cellLine(line, width)
		overlay := cellLine(layer.block.lines[y-layer.y], layer.width)
		start, end := overwriteLayerCells(row, layer.x, overlay)
		covered = append(covered, Rect{X: start, Y: y, Width: end - start, Height: 1})
		out.lines[y] = cellLineText(row)
	}
	occluded := func(r Rect) bool {
		for _, c := range covered {
			if r.X < c.X+c.Width && c.X < r.X+r.Width &&
				r.Y < c.Y+c.Height && c.Y < r.Y+r.Height {
				return true
			}
		}
		return false
	}
	// Conservative policy: a partially covered action is not actionable.
	// Drop intersecting lower paint metadata to prevent color bleeding
	// through opaque overlays. Uncovered elements retain their styling.
	for _, spot := range base.spots {
		if !occluded(Rect{X: spot.Start, Y: spot.Line, Width: spot.End - spot.Start, Height: 1}) {
			out.spots = append(out.spots, spot)
		}
	}
	for _, el := range base.elements {
		if !occluded(el.Bounds) {
			out.elements = append(out.elements, el)
		}
	}
	offsetBlock(&out, layer.block, layer.x, layer.y)
	return out
}
