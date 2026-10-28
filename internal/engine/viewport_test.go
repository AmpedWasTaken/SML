package engine

import (
	"reflect"
	"strings"
	"testing"
)

func TestWindowScrollsRenderedFrameAndClipsActions(t *testing.T) {
	view, err := Parse([]byte("<screen name=\"windows\" version=\"3\"><text>Header</text><button id=\"hidden\" label=\"Hidden\" action=\"app.hidden\"/><text>Details</text><button id=\"visible\" label=\"Visible\" action=\"app.visible\"/></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 80)
	if err != nil {
		t.Fatal(err)
	}
	original := append([]Hotspot(nil), frame.Hotspots...)
	window, err := frame.Window(0, 2, 40, 3)
	if err != nil {
		t.Fatal(err)
	}
	if window.Width != 40 || window.Height != 3 || len(window.Cells) != 3 ||
		window.Lines[0] != "Details" || window.Lines[2] != "" {
		t.Fatalf("incorrect viewport: %+v", window)
	}
	if len(window.Hotspots) != 1 {
		t.Fatalf("clipped control stayed interactive: %+v", window.Hotspots)
	}
	spot := window.Hotspots[0]
	if spot.ID != "visible" || spot.Line != 1 || spot.Start < 0 || spot.End > window.Width {
		t.Fatalf("incorrect translated hotspot: %+v", spot)
	}
	if hit, ok := window.HitTest(spot.Start, spot.Line); !ok || hit.Action != "app.visible" {
		t.Fatalf("visible control cannot be hit: %+v, ok=%v", hit, ok)
	}
	if _, ok := window.HitTest(spot.Start, 2); ok {
		t.Fatal("blank viewport row is interactive")
	}
	if !reflect.DeepEqual(frame.Hotspots, original) {
		t.Fatal("viewport changed the original frame")
	}
	if !strings.Contains(window.Lines[1], "Visible") {
		t.Fatal(window.Lines)
	}
}

func TestWindowDoesNotExposeWideGraphemeFragments(t *testing.T) {
	frame := Frame{
		Width: 4, Height: 1,
		Cells: [][]Cell{{
			{Grapheme: "A", ElementID: "control"},
			{Grapheme: "👩‍💻", ElementID: "control"},
			{Continuation: true, ElementID: "control"},
			{Grapheme: "Z", ElementID: "control"},
		}},
		Hotspots: []Hotspot{{ID: "control", Line: 0, Start: 0, End: 4, Action: "app.show"}},
		Elements: []Element{{ID: "control", Kind: "button", Bounds: Rect{Width: 4, Height: 1}}},
	}
	left, err := frame.Window(0, 0, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if left.Lines[0] != "A" || left.Cells[0][1].ElementID != "" {
		t.Fatalf("right-clipped emoji was painted or interactive: %+v", left)
	}
	if _, ok := left.HitTest(1, 0); ok {
		t.Fatal("clipped emoji must not be clickable")
	}
	emoji, err := frame.Window(1, 0, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if emoji.Lines[0] != "👩‍💻" || !emoji.Cells[0][1].Continuation {
		t.Fatalf("complete emoji was broken: %+v", emoji)
	}
	for x := 0; x < 2; x++ {
		if hit, ok := emoji.HitTest(x, 0); !ok || hit.ID != "control" {
			t.Fatalf("complete emoji cell %d not interactive: %+v, ok=%v", x, hit, ok)
		}
	}
	right, err := frame.Window(2, 0, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if right.Lines[0] != " Z" || right.Hotspots[0].Start != 1 ||
		right.Elements[0].Bounds != (Rect{Width: 2, Height: 1}) {
		t.Fatalf("orphan continuation or element geometry: %+v", right)
	}
	if _, ok := right.HitTest(0, 0); ok {
		t.Fatal("orphan continuation cell remained interactive")
	}
	if _, ok := right.HitTest(1, 0); !ok {
		t.Fatal("visible cell lost its action")
	}
	if frame.Cells[0][2].ElementID != "control" {
		t.Fatal("source cell was modified")
	}
}

func TestWindowBoundsAndEmptyViewport(t *testing.T) {
	frame := Frame{Width: 20, Height: 1, Cells: [][]Cell{make([]Cell, 20)}}
	for _, input := range [][4]int{
		{-1, 0, 10, 10}, {0, -1, 10, 10}, {0, 0, 0, 1},
		{0, 0, 241, 1}, {0, 0, 1, 0}, {0, 0, 1, 3001},
	} {
		if _, err := frame.Window(input[0], input[1], input[2], input[3]); err == nil {
			t.Fatalf("accepted invalid viewport %+v", input)
		}
	}
	offscreen, err := frame.Window(100000, 100000, 20, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(offscreen.Lines) != 3 || len(offscreen.Cells[2]) != 20 ||
		len(offscreen.Elements) != 0 || len(offscreen.Hotspots) != 0 {
		t.Fatalf("offscreen viewport was not empty: %+v", offscreen)
	}
}

func TestFrameHitTestBoundsExclusiveEndAndOverlayPrecedence(t *testing.T) {
	frame := Frame{
		Width: 4, Height: 1,
		Cells: [][]Cell{{{Grapheme: "a", ElementID: "top"}, {Grapheme: "b", ElementID: "top"}, {}, {}}},
		Hotspots: []Hotspot{
			{ID: "base", Line: 0, Start: 0, End: 2},
			{ID: "top", Line: 0, Start: 0, End: 2},
		},
	}
	for _, coordinate := range [][2]int{{-1, 0}, {0, -1}, {4, 0}, {0, 1}, {2, 0}} {
		if _, ok := frame.HitTest(coordinate[0], coordinate[1]); ok {
			t.Fatalf("out-of-range or exclusive edge hit: %+v", coordinate)
		}
	}
	if spot, ok := frame.HitTest(0, 0); !ok || spot.ID != "top" {
		t.Fatalf("last overlapping hotspot must win: %+v, %v", spot, ok)
	}
}
