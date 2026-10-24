package engine

import (
	"strings"
	"testing"
)

func TestWeightedGridKeepsPanelsVisibleForUnevenWeights(t *testing.T) {
	widths, ok := weightedGridWidths(80, 2, 3, []int{16, 1, 1})
	if !ok {
		t.Fatal("three panes should fit in 80 cells")
	}
	if widths[0]+widths[1]+widths[2] != 76 {
		t.Fatalf("did not allocate all cells: %v", widths)
	}
	for _, width := range widths {
		if width < 14 {
			t.Fatalf("pane narrower than 14 cells: %v", widths)
		}
	}
	v, err := Parse([]byte(`<screen name="uneven" version="3"><grid columns="3" weights="16,1,1" collapse-below="20" gap="2"><panel><button label="First" action="nav.first"/></panel><panel><button label="Second" action="nav.second"/></panel><panel><button label="Third" action="nav.third"/></panel></grid></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := v.RenderFrame(nil, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 3 {
		t.Fatalf("missing grid actions: %+v", frame.Hotspots)
	}
	if frame.Hotspots[0].Line != frame.Hotspots[1].Line || frame.Hotspots[1].Line != frame.Hotspots[2].Line {
		t.Fatalf("grid unexpectedly collapsed vertically: %+v", frame.Hotspots)
	}
	for _, line := range frame.Lines {
		if CellWidth(line) > 80 {
			t.Fatalf("grid line overflow: %q", line)
		}
	}
}

func TestTinyPaddedPanelClipsHotspotsToVisibleCells(t *testing.T) {
	source := `<screen name="tiny" version="3"><panel title="Tiny" padding="3"><button label="Long button to refresh" action="nav.refresh"/></panel></screen>`
	v, err := Parse([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{12, 13, 14, 16} {
		frame, err := v.RenderFrame(nil, width)
		if err != nil {
			t.Fatalf("width %d: %v", width, err)
		}
		for _, line := range frame.Lines {
			if CellWidth(line) > width {
				t.Fatalf("width %d overflow %q", width, line)
			}
		}
		for _, hit := range frame.Hotspots {
			if hit.End <= hit.Start || hit.End > width {
				t.Fatalf("width %d bad hotspot %+v", width, hit)
			}
			if hit.Line >= len(frame.Lines) {
				t.Fatal("hotspot outside rendered rows")
			}
			if frame.Cells[hit.Line][hit.Start].ElementID != hit.ID {
				t.Fatal("click target not mapped to screen cell")
			}
		}
	}
}

func TestNarrowPanelDoesNotCreateOffscreenClickTargets(t *testing.T) {
	v, err := Parse([]byte(`<screen name="tiny" version="3"><panel border="none" padding="3"><button label="Status" command="status"/></panel></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := v.RenderFrame(nil, 12)
	if err != nil {
		t.Fatal(err)
	}
	for _, hit := range frame.Hotspots {
		if hit.Start < 0 || hit.End > 12 {
			t.Fatalf("out of range click target %+v", hit)
		}
	}
	if strings.Contains(strings.Join(frame.Lines, " "), "\x1b[") {
		t.Fatal("unexpected control codes")
	}
}
