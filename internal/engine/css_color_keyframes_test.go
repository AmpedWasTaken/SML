package engine

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCSSRGBKeyframesPlayback(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		@keyframes shift {
			from { border-color: #FF0000; }
			to { border-color: #0000FF; }
		}
		panel {
			border: round;
			border-color: #111111;
			animation: shift 1s linear;
			animation-delay: 250ms;
			animation-iteration-count: 2;
			animation-direction: alternate;
		}
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><panel title="Preview"><button id="open" label="Open" action="app.open"/></panel></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	frame, err := view.RenderFrame(nil, 60)
	if err != nil || !view.HasAnimation() {
		t.Fatalf("animation missing: %v", err)
	}
	for _, tc := range []struct {
		ms    int
		color string
	}{
		{0, "17;17;17"},
		{249, "17;17;17"},
		{250, "255;0;0"},
		{1250, "0;0;255"},
		{2250, "255;0;0"},
		{9000, "255;0;0"},
	} {
		rows, spots, _ := view.RenderGlowInteractive(nil, 60, GlowOptions{TrueColor: true, Animate: true, Phase: time.Duration(tc.ms) * time.Millisecond})
		if !strings.Contains(strings.Join(rows, "\n"), "\x1b[38;2;"+tc.color+"m") || !reflect.DeepEqual(spots, frame.Hotspots) {
			t.Fatalf("incorrect playback at %dms: %q %+v", tc.ms, rows, spots)
		}
	}
	static, _ := PaintGlowFrame(frame, GlowOptions{TrueColor: true})
	reduced, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true, Animate: true, ReducedMotion: true, Phase: time.Second})
	if err != nil || !reflect.DeepEqual(static, reduced) {
		t.Fatal("reduced motion did not preserve static paint")
	}
}

func TestColorCurveMidpoints(t *testing.T) {
	_, stops, err := parseColorCurve("foreground|0:#000000,50:#FFFFFF,100:#FF0000")
	if err != nil {
		t.Fatal(err)
	}
	if got := sampleColorCurve(stops, .25); got != "#808080" {
		t.Fatalf("incorrect midpoint: %s", got)
	}
	if got := sampleColorCurve(stops, .75); got != "#ff8080" {
		t.Fatalf("incorrect midpoint: %s", got)
	}
}

func TestCSSPlaybackValidation(t *testing.T) {
	for _, source := range []string{
		"@keyframes bad { from { color: #000000; } to { background-color: #FFFFFF; } }",
		"@keyframes bad { from { color: red; } to { color: #FFFFFF; } }",
		"@keyframes bad { 50% { color: #000000; } to { color: #FFFFFF; } }",
		"panel { animation: pulse 1s; animation-delay: -1s; }",
		"panel { animation: pulse 1s; animation-direction: sideways; }",
		"panel { animation: pulse 1s; animation-iteration-count: 101; }",
		"panel { animation-delay: 1s; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><panel><text>Preview</text></panel></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("accepted invalid keyframe or playback setting: %s", source)
		}
	}
}
