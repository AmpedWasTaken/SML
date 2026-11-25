package engine

import (
	"strings"
	"testing"
)

func TestBoxOverflowScroll(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", ".scroll { overflow: auto; max-height: 2; }")
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><box class="scroll" scroll-key="state.scroll"><button id="first" label="First" action="app.first"/><button id="second" label="Second" action="app.second"/><button id="third" label="Third" action="app.third"/></box></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	for _, tc := range []struct {
		offset string
		first  string
		second string
	}{
		{"0", "first", "second"},
		{"1", "second", "third"},
		{"999", "second", "third"},
		{"-9", "first", "second"},
	} {
		values := map[string]string{"state.scroll": tc.offset}
		frame, err := view.RenderFrame(values, 60)
		if err != nil {
			t.Fatal(err)
		}
		if frame.Height != 2 || len(frame.Hotspots) != 2 {
			t.Fatalf("unexpected clipped frame: %+v", frame.Hotspots)
		}
		for i, id := range []string{tc.first, tc.second} {
			spot := frame.Hotspots[i]
			got, ok := frame.HitTest(spot.Start, i)
			if !ok || spot.ID != id || got.ID != id {
				t.Fatalf("invalid visible hit target: %+v", spot)
			}
		}
		legacy, err := view.RenderChecked(values, 60)
		if err != nil || len(legacy) != 2 {
			t.Fatalf("legacy viewport failed: %v", err)
		}
	}
	if !view.BindingDependencies().AffectedBy("state.scroll") {
		t.Fatal("missing scroll dependency")
	}
}

func TestBoxOverflowHidden(t *testing.T) {
	view, err := Parse([]byte(`<screen name="home" version="3"><box overflow="hidden" max-height="1"><text>Visible</text><button label="Hidden" action="app.open"/></box></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 0 || frame.Height != 1 || !strings.Contains(frame.Lines[0], "Visible") {
		t.Fatal("hidden content remained interactive")
	}
}

func TestBoxOverflowValidation(t *testing.T) {
	for _, attrs := range []string{`overflow="bad"`, `max-height="0"`, `scroll-key="app.offset"`} {
		if _, err := Parse([]byte(`<screen name="home" version="3"><box ` + attrs + `/></screen>`)); err == nil {
			t.Fatalf("accepted invalid box attributes %s", attrs)
		}
	}
}
