package engine

import "fmt"

// A sticky header reserves rows at the top of a scroll viewport. The
// remaining body scrolls independently; it never overlaps the header.
func (v View) renderStickyBox(n Node, values map[string]string, width, depth int, path string) (layoutBlock, error) {
	header, err := v.renderLayoutNode(n.Children[0], values, width, depth+1, fmt.Sprintf("%s/0:box", path))
	if err != nil {
		return layoutBlock{}, err
	}
	limit := boundedInt(n.Attr("max-height", "3000"), 3000, 1, 3000)
	top := boundedInt(n.Children[0].Attr("top", "0"), 0, 0, 3000)
	top = min(top, limit-1)
	visibleHeader := min(len(header.lines), limit-top)
	header = sliceLayoutRows(header, 0, visibleHeader)
	out := layoutBlock{lines: make([]string, top)}
	offsetBlock(&out, header, 0, top)
	out.lines = append(out.lines, header.lines...)
	remaining := limit - len(out.lines)
	if remaining == 0 {
		return out, nil
	}

	body := layoutBlock{}
	for i, child := range n.Children[1:] {
		part, err := v.renderLayoutNode(child, values, width, depth+1, fmt.Sprintf("%s/%d:%s", path, i+1, child.Kind))
		if err != nil {
			return layoutBlock{}, err
		}
		offsetBlock(&body, part, 0, len(body.lines))
		body.lines = append(body.lines, part.lines...)
		if len(body.lines) > 3000 {
			return layoutBlock{}, fmt.Errorf("SML sticky body exceeds row limit")
		}
	}
	attrs := make(map[string]string, len(n.Attrs))
	for key, value := range n.Attrs {
		attrs[key] = value
	}
	attrs["max-height"] = fmt.Sprint(remaining)
	n.Attrs = attrs
	body = clipBoxViewport(n, values, body)
	offsetBlock(&out, body, 0, len(out.lines))
	out.lines = append(out.lines, body.lines...)
	return out, nil
}

func hasStickyHeader(n Node) bool {
	return n.Kind == "box" && n.Attr("overflow", "visible") == "auto" &&
		len(n.Children) > 0 && n.Children[0].Kind == "box" &&
		n.Children[0].Attr("position", "static") == "sticky"
}
