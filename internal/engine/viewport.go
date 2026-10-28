package engine

import "errors"

// HitTest returns the topmost visible action at zero-based display-cell
// coordinates. Its right boundary is exclusive. The last overlapping hotspot
// wins; terminal I/O and permissions remain the host's responsibility.
func (f Frame) HitTest(x, y int) (Hotspot, bool) {
	if x < 0 || y < 0 || x >= f.Width || y >= f.Height {
		return Hotspot{}, false
	}
	for i := len(f.Hotspots) - 1; i >= 0; i-- {
		spot := f.Hotspots[i]
		if spot.Line != y || spot.End <= spot.Start || x < spot.Start || x >= spot.End {
			continue
		}
		// A clipped wide glyph may leave a blank cell inside the numerical
		// interval. Never dispatch a control from an invisible cell.
		if len(f.Cells) > 0 {
			if y >= len(f.Cells) || x >= len(f.Cells[y]) || f.Cells[y][x].ElementID != spot.ID {
				continue
			}
		}
		return spot, true
	}
	return Hotspot{}, false
}

// Window returns a fixed-size, display-cell viewport over this frame.
// scrollX/scrollY are nonnegative origin offsets into the original frame;
// content beyond its edges is blank. Hotspots and element bounds are translated
// and clipped, and orphan fragments of wide graphemes are removed. The source
// frame is never mutated. Hosts own scroll state, key events and terminal I/O.
//
// This is a post-render viewport: large datasets still require the host to
// supply a bounded data window rather than rendering unbounded records.
func (f Frame) Window(scrollX, scrollY, width, height int) (Frame, error) {
	if scrollX < 0 || scrollY < 0 || width < 1 || width > 240 || height < 1 || height > 3000 {
		return Frame{}, errors.New("SML viewport requires nonnegative offsets, width 1..240 and height 1..3000")
	}
	out := Frame{
		Width: width, Height: height,
		Lines: make([]string, height), Cells: make([][]Cell, height),
	}
	for y := 0; y < height; y++ {
		row := make([]Cell, width)
		out.Cells[y] = row
		if scrollY >= f.Height || y >= f.Height-scrollY || scrollY+y >= len(f.Cells) || scrollX >= f.Width {
			continue
		}
		source := f.Cells[scrollY+y]
		if scrollX >= len(source) {
			continue
		}
		visible := min(width, len(source)-scrollX)
		copy(row, source[scrollX:scrollX+visible])
		// A left-clipped grapheme must never leave continuation cells that
		// render without their leading character.
		for x := 0; x < width && row[x].Continuation; x++ {
			row[x] = Cell{}
		}
		for x := 0; x < width; x++ {
			if row[x].Continuation || row[x].Grapheme == "" {
				continue
			}
			span := CellWidth(row[x].Grapheme)
			valid := span >= 1 && span <= width-x
			if valid {
				for j := 1; j < span; j++ {
					if !row[x+j].Continuation {
						valid = false
						break
					}
				}
			}
			if !valid {
				for j := x; j < width && (j == x || row[j].Continuation); j++ {
					row[j] = Cell{}
				}
				continue
			}
			x += span - 1
		}
		out.Lines[y] = cellLineText(row)
	}
	for _, spot := range f.Hotspots {
		if spot.Line < scrollY || spot.Line-scrollY >= height {
			continue
		}
		y := spot.Line - scrollY
		start, end := max(0, spot.Start-scrollX), min(width, spot.End-scrollX)
		if start >= end {
			continue
		}
		// Trim zero-width clipped fragments that lost their leading
		// grapheme, but preserve normal interactive whitespace.
		for start < end && out.Cells[y][start].ElementID != spot.ID {
			start++
		}
		for end > start && out.Cells[y][end-1].ElementID != spot.ID {
			end--
		}
		if start < end {
			spot.Line, spot.Start, spot.End = y, start, end
			out.Hotspots = append(out.Hotspots, spot)
		}
	}
	for _, element := range f.Elements {
		r := element.Bounds
		if r.Width <= 0 || r.Height <= 0 || r.X+r.Width <= scrollX || r.Y+r.Height <= scrollY ||
			(r.X >= scrollX && r.X-scrollX >= width) || (r.Y >= scrollY && r.Y-scrollY >= height) {
			continue
		}
		left, top := max(0, r.X-scrollX), max(0, r.Y-scrollY)
		right, bottom := min(width, r.X+r.Width-scrollX), min(height, r.Y+r.Height-scrollY)
		if right > left && bottom > top {
			element.Bounds = Rect{X: left, Y: top, Width: right - left, Height: bottom - top}
			out.Elements = append(out.Elements, element)
		}
	}
	return out, nil
}
