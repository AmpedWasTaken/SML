package engine

import (
	"reflect"
	"strings"
	"testing"
)

func TestCSSMediaAndFlex(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		row { display: flex; gap: 2ch; justify-content: space-between; }
		.sidebar { flex-basis: 20ch; flex-grow: 0; }
		.main { flex-basis: 20ch; flex-grow: 1; }
		button { border: round; }
		@media (max-width: 60ch) {
			row { flex-direction: column; row-gap: 2; }
		}
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<row><box class="sidebar"><button id="nav" label="Navigation" action="nav.home"/></box><box class="main"><button id="main" label="Workspace" action="app.open"/></box></row>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	wide, err := view.RenderFrame(nil, 90)
	if err != nil {
		t.Fatal(err)
	}
	if len(wide.Hotspots) != 2 || wide.Hotspots[0].Line != wide.Hotspots[1].Line || wide.Hotspots[1].Start < 22 {
		t.Fatalf("unexpected horizontal flex layout: %+v", wide.Hotspots)
	}
	narrow, err := view.RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(narrow.Hotspots) != 2 || narrow.Hotspots[1].Line-narrow.Hotspots[0].Line != 5 {
		t.Fatalf("unexpected stacked layout: %+v", narrow.Hotspots)
	}
	for _, frame := range []Frame{wide, narrow} {
		for _, spot := range frame.Hotspots {
			got, ok := frame.HitTest(spot.Start, spot.Line)
			if !ok || got.ID != spot.ID {
				t.Fatalf("incorrect interactive geometry: %+v", spot)
			}
		}
		for _, line := range frame.Lines {
			if CellWidth(line) > frame.Width {
				t.Fatalf("line overflow: %q", line)
			}
		}
	}
	again, err := view.RenderFrame(nil, 90)
	if err != nil || !reflect.DeepEqual(again, wide) {
		t.Fatal("rendering a different viewport modified the source view")
	}
}

func TestCSSMediaGridTracks(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		grid { grid-template-columns: 1fr 1fr; collapse-below: 20; }
		@media (max-width: 60ch) { grid { grid-template-columns: 1fr; } }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><grid><panel><text>First</text></panel><panel><text>Second</text></panel></grid></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	wide, err := views["home"].RenderFrame(nil, 90)
	if err != nil {
		t.Fatal(err)
	}
	narrow, err := views["home"].RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	if wide.Height != 3 || narrow.Height != 6 {
		t.Fatalf("unexpected grid heights: wide=%d narrow=%d", wide.Height, narrow.Height)
	}
}

func TestCSSControlPaintStates(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		button { border: round; color: #FFFFFF; }
		button:hover { color: #00FF00; }
		button:focus { border-color: #0000FF; }
		button:disabled { color: #777777; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><button id="active" label="Active" action="app.open"/><button id="disabled" label="Disabled" action="app.open" enabled-when="app.enabled"/></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := views["home"].RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 1 || frame.Hotspots[0].ID != "active" {
		t.Fatalf("disabled control became clickable: %+v", frame.Hotspots)
	}
	painted, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true, FocusedID: "active", HoveredID: "active"})
	if err != nil {
		t.Fatal(err)
	}
	output := strings.Join(painted, "\n")
	for _, sgr := range []string{"\x1b[38;2;0;255;0m", "\x1b[38;2;0;0;255m", "\x1b[38;2;119;119;119m"} {
		if !strings.Contains(output, sgr) {
			t.Fatalf("missing state color %q: %q", sgr, output)
		}
	}
	plain, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true, NoColor: true, FocusedID: "active"})
	if err != nil || !reflect.DeepEqual(plain, frame.Lines) {
		t.Fatal("state changed plain frame")
	}
}

func TestCSSMediaAndStateRejectBehaviorChanges(t *testing.T) {
	for _, source := range []string{
		"@media (max-width: 60ch) { grid { columns: 1; } }",
		"@media (min-height: 60ch) { row { gap: 1; } }",
		"button:focus { action: app.open; }",
		"button:hover { border: double; }",
		"row { display: absolute; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><row><button label="Open" action="app.open"/></row></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("invalid rule accepted: %s", source)
		}
	}
}
