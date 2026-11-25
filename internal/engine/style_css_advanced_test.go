package engine

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCSSAnimationKeyframes(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		@keyframes breathe {
			from { brightness: 40%; }
			50% { brightness: 100%; }
			to { brightness: 40%; }
		}
		panel { border: double; glow: medium; }
		panel { animation: breathe 2s linear infinite; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><panel title="Health"><button label="Status" command="status"/></panel></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	frame, err := view.RenderFrame(nil, 60)
	if err != nil || !view.HasAnimation() {
		t.Fatalf("missing animation: %v", err)
	}
	if len(frame.Hotspots) != 1 {
		t.Fatalf("missing hotspot: %+v", frame.Hotspots)
	}
	a, spots, _ := view.RenderGlowInteractive(nil, 60, GlowOptions{TrueColor: true, Animate: true})
	b, spots2, _ := view.RenderGlowInteractive(nil, 60, GlowOptions{TrueColor: true, Animate: true, Phase: time.Second})
	if reflect.DeepEqual(a, b) || !reflect.DeepEqual(spots, spots2) || !reflect.DeepEqual(spots, frame.Hotspots) {
		t.Fatal("animation did not change paint, or altered interaction geometry")
	}
	plain, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true, NoColor: true, Animate: true})
	if err != nil || !reflect.DeepEqual(plain, frame.Lines) {
		t.Fatal("plain output changed")
	}
	static, _ := PaintGlowFrame(frame, GlowOptions{TrueColor: true})
	reduced, _ := PaintGlowFrame(frame, GlowOptions{TrueColor: true, Animate: true, ReducedMotion: true, Phase: time.Second})
	if !reflect.DeepEqual(static, reduced) {
		t.Fatal("reduced motion changed static output")
	}
}

func TestSMLCSSResponsiveLayoutAndHotspots(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		:root { --edge: round; }
		panel, button { border: var(--edge); }
		box { width: 80%; min-width: 20ch; max-width: 60ch; }
		grid { grid-template-columns: 1fr 2fr; column-gap: 2ch; row-gap: 2; collapse-below: 55ch; align-items: end; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><box align="center"><grid><panel title="One"><text>First</text></panel><button id="second" label="Second" command="status"/><button id="third" label="Third" command="help"/></grid></box></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	for _, width := range []int{32, 80, 120} {
		frame, err := view.RenderFrame(nil, width)
		if err != nil || len(frame.Hotspots) != 2 {
			t.Fatalf("%d: %v %+v", width, err, frame.Hotspots)
		}
		for _, spot := range frame.Hotspots {
			if got, ok := frame.HitTest(spot.Start, spot.Line); !ok || got.ID != spot.ID {
				t.Fatalf("incorrect hotspot: %+v", spot)
			}
			if !strings.Contains(frame.Lines[spot.Line], spot.Label) && width >= 80 {
				t.Fatalf("hitbox not on rendered control: %+v", spot)
			}
		}
		for _, line := range frame.Lines {
			if CellWidth(line) > width {
				t.Fatalf("overflow at %d: %q", width, line)
			}
		}
		boxWidth := 0
		for _, element := range frame.Elements {
			if element.Kind == "box" {
				boxWidth = element.Bounds.Width
			}
		}
		want := min(width, max(20, min(width*80/100, 60)))
		if boxWidth != want {
			t.Fatalf("%d: box width %d, want %d", width, boxWidth, want)
		}
		if width >= 80 && frame.Hotspots[1].Line-frame.Hotspots[0].Line < 3 {
			t.Fatal("missing grid row gap")
		}
	}
}

func TestCSSAdvancedValidation(t *testing.T) {
	for _, source := range []string{
		"@keyframes bad { from { brightness: 0%; } to { brightness: 100%; } }",
		"@keyframes bad { from { brightness: 40%; } }",
		"panel { animation: missing 2s linear infinite; }",
		"panel { animation: pulse 1ms; }",
		"button { width: 30%; }",
		"grid { grid-template-columns: 0fr 1fr; }",
		"grid { row-gap: 99; }",
		"button:hover { border: round; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><panel><button label="Status" command="status"/></panel><grid><text>Test</text></grid></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("accepted %s", source)
		}
	}
}
