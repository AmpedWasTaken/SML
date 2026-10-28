package engine

import (
	"errors"
	"reflect"
)

// CellRun replaces the cells in one row from Start (inclusive) to
// Start+len(Cells) (exclusive). A blank Cell clears prior painted content.
// A run always includes complete grapheme clusters, including the continuation
// cells of double-width characters. Its Cells slice belongs to the diff.
type CellRun struct {
	Row   int
	Start int
	Cells []Cell
}

// FrameDiff describes changes to a terminal-cell snapshot without generating
// ANSI, moving a cursor or dispatching an action.
type FrameDiff struct {
	Width           int
	Height          int
	FullRedraw      bool
	Runs            []CellRun
	HotspotsChanged bool
	ElementsChanged bool
}

// DiffFrames compares two cell snapshots. A missing previous frame or a
// geometry change requests a full redraw. Otherwise changes are grouped into
// minimal contiguous row runs and expanded to whole grapheme clusters.
// Changes to interaction/layout metadata are reported separately: modifying an
// action never requires repainting the text when the glyphs stay identical.
// A host must still decide when to redraw for terminal capability changes,
// cursor movement, overlays and styles not represented in Frame.Cells.
func DiffFrames(previous, current Frame) (FrameDiff, error) {
	if err := validateDiffFrame(current); err != nil {
		return FrameDiff{}, err
	}
	return computeFrameDiff(previous, current)
}

func validateDiffFrame(f Frame) error {
	if f.Width < 1 || f.Width > 240 || f.Height < 0 || f.Height > 3000 || len(f.Cells) != f.Height ||
		(len(f.Lines) != 0 && len(f.Lines) != f.Height) {
		return errors.New("invalid SML frame shape for diff")
	}
	for _, row := range f.Cells {
		if len(row) != f.Width {
			return errors.New("invalid SML frame cell row for diff")
		}
	}
	return nil
}

func computeFrameDiff(previous, current Frame) (FrameDiff, error) {
	out := FrameDiff{
		Width: current.Width, Height: current.Height,
		HotspotsChanged: !reflect.DeepEqual(previous.Hotspots, current.Hotspots),
		ElementsChanged: !reflect.DeepEqual(previous.Elements, current.Elements),
	}
	if err := validateDiffFrame(previous); err != nil || previous.Width != current.Width || previous.Height != current.Height {
		out.FullRedraw = true
		for y, row := range current.Cells {
			out.Runs = append(out.Runs, CellRun{Row: y, Cells: append([]Cell(nil), row...)})
		}
		return out, nil
	}
	// Reuse the scratch mask across rows so one changed metric does not
	// allocate for every otherwise-unchanged row in a dashboard.
	dirty := make([]bool, current.Width)
	for y, row := range current.Cells {
		clear(dirty)
		old := previous.Cells[y]
		for x, cell := range row {
			if cell.Grapheme != old[x].Grapheme || cell.Continuation != old[x].Continuation {
				// Replacing half of a wide emoji can corrupt the neighbor.
				// Invalidate its full span in BOTH the old and new frames.
				markClusterDirty(dirty, row, x)
				markClusterDirty(dirty, old, x)
			}
		}
		if len(previous.Lines) != 0 && len(current.Lines) != 0 && previous.Lines[y] != current.Lines[y] &&
			(previous.Lines[y] != cellLineText(old) || current.Lines[y] != cellLineText(row)) {
			// A future styled frame may have invisible ANSI changes which
			// are not yet represented in Cell. Fall back to the entire line.
			for x := range dirty {
				dirty[x] = true
			}
		}
		for x := 0; x < current.Width; {
			if !dirty[x] {
				x++
				continue
			}
			start := x
			for x < current.Width && dirty[x] {
				x++
			}
			out.Runs = append(out.Runs, CellRun{Row: y, Start: start, Cells: append([]Cell(nil), row[start:x]...)})
		}
	}
	return out, nil
}

func markClusterDirty(dirty []bool, row []Cell, x int) {
	left := x
	for left > 0 && row[left].Continuation {
		left--
	}
	right := max(x+1, left+1)
	for right < len(row) && row[right].Continuation {
		right++
	}
	for i := left; i < right; i++ {
		dirty[i] = true
	}
}
