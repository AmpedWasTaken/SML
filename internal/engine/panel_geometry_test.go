package engine

import (
	"strings"
	"testing"
)

func TestPanelAlignedColorBounds(t *testing.T) {
	for _, align := range []string{"center", "right"} {
		view, err := Parse([]byte(`<screen name="sample" version="3"><panel align="` + align + `" padding="1" border="round"><text foreground="#ff0000">Hello</text><button label="Next" action="sample.next" id="next"/></panel></screen>`))
		if err != nil {
			t.Fatal(err)
		}
		frame, err := view.RenderFrame(nil, 40)
		if err != nil {
			t.Fatal(err)
		}
		painted, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(painted[1], "\x1b[38;2;255;0;0mHello") {
			t.Fatalf("aligned text did not receive its color: %q", painted[1])
		}
		spot := frame.Hotspots[0]
		if got, ok := frame.HitTest(spot.Start, spot.Line); !ok || got.ID != "next" {
			t.Fatal("aligned control lost its hotspot")
		}
	}
}

func TestNarrowPanelClipsChildBounds(t *testing.T) {
	view, err := Parse([]byte(`<screen name="sample" version="3"><panel padding="3"><panel glow="soft" border="round" padding="0"><text>Content</text></panel></panel></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 12)
	if err != nil {
		t.Fatal(err)
	}
	for _, element := range frame.Elements {
		r := element.Bounds
		if r.X < 0 || r.Y < 0 || r.X+r.Width > frame.Width || r.Y+r.Height > frame.Height {
			t.Fatalf("out-of-frame element: %+v", element)
		}
	}
	if _, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true}); err != nil {
		t.Fatalf("narrow panel failed to paint: %v", err)
	}
}

func TestBorderlessPanelEmptyTitleDoesNotShiftControl(t *testing.T) {
	view, err := Parse([]byte(`<screen name="sample" version="3"><panel border="none" title="{{sample.title}}"><button label="Next" action="sample.next" id="next"/></panel></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"", "Title"} {
		frame, err := view.RenderFrame(map[string]string{"sample.title": title}, 40)
		if err != nil {
			t.Fatal(err)
		}
		spot := frame.Hotspots[0]
		want := 0
		if title != "" {
			want = 1
		}
		if spot.Line != want {
			t.Fatalf("title %q: hotspot row %d, want %d", title, spot.Line, want)
		}
	}
}
