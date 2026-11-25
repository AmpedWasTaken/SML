package engine

import (
	"reflect"
	"strings"
	"testing"
)

func TestCSSPositionedLayers(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		.stage { position: relative; }
		.front { position: absolute; left: 0ch; top: 0; width: 24ch; z-index: 10; }
		.back { position: absolute; left: 0; top: 0; width: 24; z-index: 5; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<box class="stage">
			<button id="base" label="Base" action="app.base"/>
			<button id="outside" label="Outside" action="app.outside"/>
			<box class="front"><button id="front" label="Front" action="app.front"/></box>
			<box class="back"><button id="back" label="Back" action="app.back"/></box>
		</box>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	for _, width := range []int{12, 40, 80} {
		frame, err := view.RenderFrame(nil, width)
		if err != nil {
			t.Fatal(err)
		}
		if frame.Height != 2 || len(frame.Hotspots) != 2 {
			t.Fatalf("unexpected frame at %d: %+v", width, frame.Hotspots)
		}
		if !strings.Contains(frame.Lines[0], "Front") || strings.Contains(frame.Lines[0], "Back") || strings.Contains(frame.Lines[0], "Base") {
			t.Fatalf("unexpected layer order: %q", frame.Lines)
		}
		for _, spot := range frame.Hotspots {
			got, ok := frame.HitTest(spot.Start, spot.Line)
			if !ok || got.ID != spot.ID || (spot.ID != "outside" && spot.ID != "front") {
				t.Fatalf("occluded or incorrect hit target: %+v", spot)
			}
		}
		legacy, err := view.RenderChecked(nil, width)
		if err != nil || !reflect.DeepEqual(legacy, frame.Lines) {
			t.Fatalf("legacy positioning mismatch: %v", err)
		}
	}
}

func TestLayerDoesNotLeaveHalfGrapheme(t *testing.T) {
	row := cellLine("A界B界Z", 12)
	start, end := overwriteLayerCells(row, 2, cellLine("XYZ", 3))
	if start != 1 || end != 6 || row[1].Grapheme != "" || row[5].Continuation {
		t.Fatalf("wide grapheme was only partially cleared: %+v (%d..%d)", row, start, end)
	}
	if got := cellLineText(row); !strings.HasPrefix(got, "A XYZ Z") {
		t.Fatalf("unexpected composed text: %q", got)
	}
}

func TestPositionedLayoutRejectsInvalidRules(t *testing.T) {
	for _, body := range []string{
		`<box position="absolute"/>`,
		`<box left="1"/>`,
		`<box position="relative"><box position="absolute" z-index="101"/></box>`,
		`<box position="relative"><box position="absolute" top="-1"/></box>`,
		`<box position="fixed"/>`,
	} {
		if _, err := Parse([]byte(`<screen name="bad" version="3">` + body + `</screen>`)); err == nil {
			t.Fatalf("invalid positioned layout accepted: %s", body)
		}
	}
}
