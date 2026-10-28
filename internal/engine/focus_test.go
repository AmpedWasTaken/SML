package engine

import (
	"reflect"
	"testing"
)

func TestFocusOrderExcludesDisabledAndPreservesRepeatOrder(t *testing.T) {
	view, err := Parse([]byte("<screen name=\"focus\" version=\"3\"><button id=\"first\" label=\"First\" action=\"app.first\"/><button id=\"disabled\" label=\"Disabled\" action=\"app.disabled\" enabled-when=\"ready\"/><repeat source=\"items\" as=\"item\" limit=\"3\"><button id=\"inspect\" label=\"Inspect\" action=\"app.inspect\"/></repeat></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{"ready": "false", "items": "one,two"}
	frame, err := view.RenderFrame(values, 60)
	if err != nil {
		t.Fatal(err)
	}
	ordered := frame.FocusOrder()
	if len(ordered) != 3 || ordered[0].ID != "first" ||
		ordered[1].ID == ordered[2].ID || ordered[1].Action != "app.inspect" ||
		ordered[2].Action != "app.inspect" {
		t.Fatalf("wrong focus order or missing repeat identities: %+v", ordered)
	}
	for _, spot := range ordered {
		if spot.ID == "disabled" {
			t.Fatal("disabled control was focusable")
		}
	}
	again, err := view.RenderFrame(values, 60)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ordered, again.FocusOrder()) {
		t.Fatal("identical renders changed keyboard focus identity")
	}
	if next, ok := frame.NextFocus("first", false); !ok || next.ID != ordered[1].ID {
		t.Fatalf("forward focus did not follow document order: %+v %v", next, ok)
	}
	if next, ok := frame.NextFocus(ordered[2].ID, false); !ok || next.ID != "first" {
		t.Fatalf("forward focus did not wrap: %+v %v", next, ok)
	}
	if previous, ok := frame.NextFocus("first", true); !ok || previous.ID != ordered[2].ID {
		t.Fatalf("backward focus did not wrap: %+v %v", previous, ok)
	}
	if fallback, ok := frame.NextFocus("unmounted", true); !ok || fallback.ID != ordered[2].ID {
		t.Fatalf("lost focus did not fall back: %+v %v", fallback, ok)
	}
}

func TestFocusOrderExcludesUnpaintedAndDuplicates(t *testing.T) {
	frame := Frame{
		Width: 6, Height: 1,
		Cells: [][]Cell{{
			{Grapheme: "A", ElementID: "visible"},
			{Grapheme: "B", ElementID: "visible"},
			{}, {}, {}, {},
		}},
		Hotspots: []Hotspot{
			{ID: "hidden", Line: 0, Start: 2, End: 5},
			{ID: "visible", Line: 0, Start: 0, End: 2, Options: []string{"one"}},
			{ID: "visible", Line: 0, Start: 0, End: 2},
			{ID: "outside", Line: 4, Start: 0, End: 1},
		},
	}
	ordered := frame.FocusOrder()
	if len(ordered) != 1 || ordered[0].ID != "visible" {
		t.Fatalf("unpainted or duplicate control was focusable: %+v", ordered)
	}
	ordered[0].Options[0] = "changed"
	if frame.Hotspots[1].Options[0] != "one" {
		t.Fatal("focus metadata exposed a writable reference to the frame")
	}
	empty := Frame{Width: 10, Height: 2}
	if _, ok := empty.NextFocus("", false); ok {
		t.Fatal("empty frame has a focusable control")
	}
	if _, ok := empty.NextFocus("", true); ok {
		t.Fatal("empty frame has backward focus")
	}
}
