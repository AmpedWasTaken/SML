package engine

import (
	"strings"
	"testing"
)

func TestWeightedGridAllocatesResponsivePanesAndHotspots(t *testing.T) {
	view, err := Parse([]byte(`<screen name="split" version="3">
	  <grid columns="2" weights="1,3" gap="2" collapse-below="20">
	    <panel title="Navigation"><button id="nav" label="Browse" action="nav.browse"/></panel>
	    <panel title="Content"><button id="main" label="Open" action="main.open"/></panel>
	  </grid>
	</screen>`))
	if err != nil {
		t.Fatal(err)
	}
	wide, err := view.RenderFrame(nil, 80)
	if err != nil {
		t.Fatal(err)
	}
	var panels []Rect
	for _, el := range wide.Elements {
		if el.Kind == "panel" {
			panels = append(panels, el.Bounds)
		}
	}
	if len(panels) != 2 || panels[0].Width != 19 || panels[1].Width != 59 || panels[1].X != 21 {
		t.Fatalf("weighted pane geometry incorrect: %+v", panels)
	}
	if len(wide.Hotspots) != 2 || wide.Hotspots[0].Line != wide.Hotspots[1].Line ||
		wide.Hotspots[0].End > wide.Hotspots[1].Start {
		t.Fatalf("weighted pane focus geometry incorrect: %+v", wide.Hotspots)
	}
	for _, hit := range wide.Hotspots {
		if wide.Cells[hit.Line][hit.Start].ElementID != hit.ID {
			t.Fatalf("clickable cell missing for %+v", hit)
		}
	}
	for _, line := range wide.Lines {
		if CellWidth(line) > 80 {
			t.Fatalf("line overflow: %q", line)
		}
	}
	narrow, err := view.RenderFrame(nil, 28)
	if err != nil {
		t.Fatal(err)
	}
	if len(narrow.Hotspots) != 2 || narrow.Hotspots[0].Line >= narrow.Hotspots[1].Line {
		t.Fatalf("narrow grid did not stack: %+v", narrow.Hotspots)
	}
	for _, line := range narrow.Lines {
		if CellWidth(line) > 28 {
			t.Fatalf("narrow line overflow: %q", line)
		}
	}
}

func TestGridWeightsRejectMalformedOrUnsafeValues(t *testing.T) {
	for _, inner := range []string{
		`<grid columns="2" weights="1"><text>Only one weight</text></grid>`,
		`<grid columns="2" weights="0,1"><text>Zero</text></grid>`,
		`<grid columns="2" weights="1,17"><text>Oversized</text></grid>`,
		`<grid columns="2" weights="1,nope"><text>Text</text></grid>`,
		`<row weights="1,1"><text>Not a grid</text></row>`,
	} {
		src := `<screen name="bad" version="3">` + inner + `</screen>`
		if _, err := Parse([]byte(src)); err == nil {
			t.Errorf("unexpectedly accepted invalid weights %s", inner)
		}
	}
	if _, err := Parse([]byte(`<screen name="bad"><grid columns="2" weights="1,1"/></screen>`)); err == nil {
		t.Fatal("weighted grid should require SML v3")
	}
}

func TestWeightedGridNarrowTwelveCellFallback(t *testing.T) {
	view, err := Parse([]byte(`<screen name="tiny" version="3"><grid columns="2" weights="1,3" collapse-below="20"><text>A</text><text>B</text></grid></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 12)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(frame.Lines, " "), "A B") {
		t.Fatalf("12-cell fallback lost content: %+v", frame.Lines)
	}
}
