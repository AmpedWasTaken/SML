package engine

// panelLineShift follows stylePanel's per-row alignment exactly. The visible
// width is measured after clipping, using display cells rather than bytes.
func panelLineShift(line string, width int, align string) int {
	slack := max(0, width-CellWidth(ClipCells(line, max(0, width))))
	switch align {
	case "center":
		return slack / 2
	case "right":
		return slack
	default:
		return 0
	}
}

// panelVisibleBounds returns the rectangular union of the visible cells of
// a child element. For a multi-line element in a per-line aligned panel the
// union may contain whitespace, but never includes the enclosing border or
// cells clipped from the child. Empty geometry is omitted from the frame.
func panelVisibleBounds(bounds Rect, lines []string, width int, align string) (Rect, bool) {
	top, bottom := max(0, bounds.Y), min(len(lines), bounds.Y+bounds.Height)
	left, right := width, 0
	first, last := bottom, top
	for y := top; y < bottom; y++ {
		shift := panelLineShift(lines[y], width, align)
		x0 := max(0, min(width, bounds.X+shift))
		x1 := max(0, min(width, bounds.X+bounds.Width+shift))
		if x1 <= x0 {
			continue
		}
		left, right = min(left, x0), max(right, x1)
		first, last = min(first, y), max(last, y+1)
	}
	if left >= right || first >= last {
		return Rect{}, false
	}
	return Rect{X: left, Y: first, Width: right-left, Height: last-first}, true
}
