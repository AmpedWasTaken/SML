package sml_test

import (
	"strings"
	"testing"

	sml "github.com/AmpedWasTaken/SML"
)

func TestPublicFacadePreservesRenderingAndActions(t *testing.T) {
	view, err := sml.Parse([]byte(`<screen name="public" version="3">
		<panel title="Status">
			<text>Hello {{user.name}}</text>
			<button label="Refresh" id="refresh" action="app.refresh"/>
		</panel>
	</screen>`))
	if err != nil {
		t.Fatal(err)
	}

	var _ sml.View = view
	lines, err := view.RenderChecked(map[string]string{"user.name": "SML"}, 60)
	if err != nil || !strings.Contains(strings.Join(lines, "\n"), "Hello SML") {
		t.Fatalf("render: %v %q", err, lines)
	}
	frame, err := view.RenderFrame(map[string]string{"user.name": "SML"}, 60)
	if err != nil {
		t.Fatal(err)
	}
	var _ sml.Frame = frame
	if len(frame.Cells) == 0 || len(frame.Hotspots) != 1 || frame.Hotspots[0].Action != "app.refresh" {
		t.Fatalf("frame unexpectedly changed: %+v", frame.Hotspots)
	}
	var _ sml.Cell = frame.Cells[0][0]
	var _ sml.Hotspot = frame.Hotspots[0]

	_, interactive, _ := view.RenderInteractive(nil, 60)
	if len(interactive) != 1 || interactive[0].ID != "refresh" {
		t.Fatalf("interactive actions unexpectedly changed: %+v", interactive)
	}
	if view.HasAnimation() {
		t.Fatal("unexpected animation")
	}
}

func TestPublicFacadeHelpers(t *testing.T) {
	bindings, err := sml.ParseActionBindings("item:state.selection")
	if err != nil || bindings["item"] != "state.selection" {
		t.Fatalf("bindings: %v %v", bindings, err)
	}
	if sml.CellWidth("界") != 2 || sml.ClipCells("界X", 2) != "界" {
		t.Fatal("grapheme width or clipping incompatible")
	}
	if sml.PadCells("界", 4) != "界  " || len(sml.WrapText("hello world", 5)) < 2 {
		t.Fatal("padding or text wrapping incompatible")
	}
	if _, err := sml.Parse([]byte(`<screen name="ok"><unsafe/></screen>`)); err == nil {
		t.Fatal("unknown elements must still fail validation")
	}
}

func TestPublicFrameViewportAndHitTestCompatibility(t *testing.T) {
	view, err := sml.Parse([]byte("<screen name=\"api\" version=\"3\"><text>Heading</text><button id=\"select\" label=\"Select\" action=\"app.select\"/></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 40)
	if err != nil {
		t.Fatal(err)
	}
	window, err := frame.Window(0, 1, 20, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(window.Hotspots) != 1 || window.Hotspots[0].Line != 0 {
		t.Fatalf("unexpected viewport metadata: %+v", window.Hotspots)
	}
	x := window.Hotspots[0].Start
	spot, ok := window.HitTest(x, 0)
	if !ok || spot.ID != "select" || spot.Action != "app.select" {
		t.Fatalf("public Frame.HitTest was not preserved: %+v, %v", spot, ok)
	}
	if _, ok := window.HitTest(x, 1); ok {
		t.Fatal("offscreen control was hittable")
	}
	var _ sml.Frame = window
	var _ sml.Hotspot = spot
}
