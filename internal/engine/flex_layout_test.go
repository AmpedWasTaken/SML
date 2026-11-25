package engine

import (
	"reflect"
	"testing"
)

func TestFlexShrinkDistribution(t *testing.T) {
	box := func(base, shrink, floor string) Node {
		return Node{Kind: "box", Attrs: map[string]string{
			"flex-basis": base, "flex-shrink": shrink, "min-width": floor,
		}}
	}
	cases := []struct {
		children []Node
		width    int
		want     []int
		fits     bool
	}{
		{[]Node{box("32", "1", "14"), box("32", "1", "14")}, 60, []int{29, 29}, true},
		{[]Node{box("32", "0", "14"), box("32", "1", "14")}, 60, []int{32, 26}, true},
		{[]Node{box("32", "1", "30"), box("32", "1", "14")}, 60, []int{30, 28}, true},
		{[]Node{box("32", "0", "14"), box("32", "0", "14")}, 60, nil, false},
		{[]Node{box("32", "1", "30"), box("32", "1", "30")}, 60, nil, false},
	}
	for _, tc := range cases {
		got, free, fits := flexWidths(tc.children, tc.width, 2)
		if fits != tc.fits || !reflect.DeepEqual(got, tc.want) || free != 0 {
			t.Fatalf("width=%d: got %v free=%d fits=%v; want %v fits=%v", tc.width, got, free, fits, tc.want, tc.fits)
		}
	}
}

func TestCSSFlexWrapPreservesControlIDs(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		row { display: flex; flex-wrap: wrap; gap: 2; row-gap: 1; collapse-below: 20; }
		box { flex-basis: 30ch; flex-grow: 0; flex-shrink: 1; }
		button { border: round; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<row>
			<box><button id="first" label="First" action="app.first"/></box>
			<box><button id="second" label="Second" action="app.second"/></box>
			<box><button id="third" label="Third" action="app.third"/></box>
		</row>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	for _, tc := range []struct {
		width int
		lines []int
	}{
		{100, []int{1, 1, 1}},
		{70, []int{1, 1, 5}},
		{32, []int{1, 5, 9}},
		{12, []int{1, 5, 9}},
	} {
		frame, err := view.RenderFrame(nil, tc.width)
		if err != nil {
			t.Fatal(err)
		}
		if len(frame.Hotspots) != 3 {
			t.Fatalf("missing controls at width %d: %+v", tc.width, frame.Hotspots)
		}
		for i, id := range []string{"first", "second", "third"} {
			spot := frame.Hotspots[i]
			got, ok := frame.HitTest(spot.Start, spot.Line)
			if spot.ID != id || spot.Line != tc.lines[i] || !ok || got.ID != id {
				t.Fatalf("width %d: unexpected hotspot %+v", tc.width, spot)
			}
		}
		for _, line := range frame.Lines {
			if CellWidth(line) > tc.width {
				t.Fatalf("overflow at %d: %q", tc.width, line)
			}
		}
	}
}

func TestFlexNewPropertiesValidation(t *testing.T) {
	for _, source := range []string{
		"row { flex-wrap: reverse; }",
		"box { flex-shrink: -1; }",
		"box { flex-shrink: 17; }",
		"button { flex-wrap: wrap; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><row><box><button label="Open" action="app.open"/></box></row></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("invalid CSS accepted: %s", source)
		}
	}
}
