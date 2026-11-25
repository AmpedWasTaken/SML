package engine

import (
	"strings"
	"testing"
)

func TestASCIIFramePreservesCellGeometry(t *testing.T) {
	view, err := Parse([]byte(`<screen name="demo" version="3"><panel title="Hello"><text>界 😀</text><button id="open" label="Open" action="open"/></panel></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 40)
	if err != nil {
		t.Fatal(err)
	}
	ascii := frame.ASCII()
	if ascii.Width != frame.Width || ascii.Height != frame.Height || len(ascii.Hotspots) != len(frame.Hotspots) || len(ascii.Hotspots) == 0 {
		t.Fatal("fallback changed geometry")
	}
	for y, line := range ascii.Lines {
		for _, r := range line {
			if r < 32 || r > 126 {
				t.Fatalf("non-ASCII output on row %d: %q", y, line)
			}
		}
		if CellWidth(line) != CellWidth(frame.Lines[y]) {
			t.Fatalf("changed cell width on row %d", y)
		}
	}
	for _, spot := range frame.Hotspots {
		if _, ok := frame.HitTest(spot.Start, spot.Line); ok {
			if _, ok := ascii.HitTest(spot.Start, spot.Line); !ok {
				t.Fatal("fallback lost hotspot")
			}
		}
	}
	if !strings.Contains(strings.Join(frame.Lines, "\n"), "界") {
		t.Fatal("source frame was mutated")
	}
}
