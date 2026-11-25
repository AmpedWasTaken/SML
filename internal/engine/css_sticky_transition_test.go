package engine

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCSSStickyHeader(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		.viewport { overflow: auto; max-height: 4; }
		.header { position: sticky; top: 0; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<box class="viewport" scroll-key="state.scroll">
			<box class="header"><button id="header" label="Header" action="app.header"/></box>
			<button id="first" label="First" action="app.first"/>
			<button id="second" label="Second" action="app.second"/>
			<button id="third" label="Third" action="app.third"/>
			<button id="fourth" label="Fourth" action="app.fourth"/>
		</box>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	view := views["home"]
	for _, tc := range []struct {
		offset string
		want   []string
	}{
		{"0", []string{"header", "first", "second", "third"}},
		{"1", []string{"header", "second", "third", "fourth"}},
		{"999", []string{"header", "second", "third", "fourth"}},
	} {
		values := map[string]string{"state.scroll": tc.offset}
		frame, err := view.RenderFrame(values, 60)
		if err != nil {
			t.Fatal(err)
		}
		if frame.Height != 4 || len(frame.Hotspots) != 4 {
			t.Fatalf("incorrect sticky viewport: %+v", frame.Hotspots)
		}
		for i, id := range tc.want {
			spot := frame.Hotspots[i]
			got, ok := frame.HitTest(spot.Start, i)
			if !ok || got.ID != id || spot.ID != id || spot.Line != i {
				t.Fatalf("unexpected visible target: %+v", spot)
			}
		}
		legacy, err := view.RenderChecked(values, 60)
		if err != nil || !reflect.DeepEqual(legacy, frame.Lines) {
			t.Fatalf("legacy sticky rendering mismatch: %v", err)
		}
	}
}

func TestCSSStickyValidation(t *testing.T) {
	for _, body := range []string{
		`<box position="sticky"/>`,
		`<box><box position="sticky"/></box>`,
		`<box overflow="auto"><text>First</text><box position="sticky"/></box>`,
		`<box overflow="auto"><box position="sticky" left="2"/></box>`,
	} {
		if _, err := Parse([]byte(`<screen name="bad" version="3">` + body + `</screen>`)); err == nil {
			t.Fatalf("invalid sticky box accepted: %s", body)
		}
	}
}

func TestCSSColorTransition(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		button { color: #000000; transition: color 200ms linear; }
		button:hover { color: #FFFFFF; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><button id="open" label="Open" action="app.open"/></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := views["home"].RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	opts := GlowOptions{TrueColor: true, Transition: true, HoveredID: "open"}
	for _, tc := range []struct {
		elapsed time.Duration
		color   string
	}{
		{0, "0;0;0"},
		{100 * time.Millisecond, "128;128;128"},
		{200 * time.Millisecond, "255;255;255"},
	} {
		opts.TransitionElapsed = tc.elapsed
		painted, err := PaintGlowFrame(frame, opts)
		if err != nil || !strings.Contains(strings.Join(painted, "\n"), "\x1b[38;2;"+tc.color+"m") {
			t.Fatalf("missing interpolated color at %v: %v %q", tc.elapsed, err, painted)
		}
	}
	opts.TransitionElapsed = 0
	opts.ReducedMotion = true
	painted, err := PaintGlowFrame(frame, opts)
	if err != nil || !strings.Contains(strings.Join(painted, "\n"), "\x1b[38;2;255;255;255m") {
		t.Fatal("reduced motion did not skip transition")
	}
	opts.ReducedMotion = false
	opts.PreviousHoveredID, opts.HoveredID = "open", ""
	opts.TransitionElapsed = 100 * time.Millisecond
	painted, err = PaintGlowFrame(frame, opts)
	if err != nil || !strings.Contains(strings.Join(painted, "\n"), "\x1b[38;2;128;128;128m") {
		t.Fatal("hover exit did not interpolate")
	}
}

func TestCSSTransitionValidation(t *testing.T) {
	for _, source := range []string{
		"button { transition: width 200ms; }",
		"button { transition: color -1s; }",
		"button { transition: color 200ms bogus; }",
		"panel { transition: color 200ms; }",
		"button:hover { transition: color 200ms; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><button label="Open" action="app.open"/></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("invalid transition accepted: %s", source)
		}
	}
}
