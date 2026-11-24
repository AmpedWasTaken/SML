package premiumwidgets

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestCatalogAndBounds(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range Catalog() {
		if seen[d.ID] {
			t.Fatal("duplicate", d.ID)
		}
		seen[d.ID] = true
		for _, w := range []int{0, 16, 32, 72, 200} {
			for _, ascii := range []bool{true, false} {
				f, err := Render(d.ID, Input{Width: w, ASCII: ascii, Value: math.NaN(), Secondary: math.Inf(1), Series: []float64{-50, 50, 200}, Phase: time.Second, Label: "\x1b[2J\nExample"})
				if err != nil {
					t.Fatal(err)
				}
				for _, line := range f.Lines {
					if len([]rune(line)) != max(16, min(w, 72)) || strings.ContainsAny(line, "\x1b\n\r") {
						t.Fatalf("%s invalid frame %q", d.ID, line)
					}
					if ascii {
						for _, r := range line {
							if r > 126 {
								t.Fatalf("%s non-ASCII %q", d.ID, line)
							}
						}
					}
				}
			}
		}
	}
	if len(seen) != 32 {
		t.Fatalf("expected 32 widgets, got %d", len(seen))
	}
	if _, err := Render("unknown", Input{}); err == nil {
		t.Fatal("unknown widget accepted")
	}
}

func TestAnimationsAndLocalControls(t *testing.T) {
	for _, id := range []string{"orbit-spinner", "pulse-beacon", "skeleton-loader", "equalizer", "shimmer-text"} {
		in := Input{Width: 32, Value: 60}
		a, _ := Render(id, in)
		in.Phase = 250 * time.Millisecond
		b, _ := Render(id, in)
		if a == b {
			t.Fatalf("%s did not animate", id)
		}
		in.ReducedMotion = true
		c, _ := Render(id, in)
		if a != c {
			t.Fatalf("%s reduced motion failed", id)
		}
	}
	in, err := Apply("slider-control", "next", Input{Selection: 4})
	if err != nil || in.Selection != 4 {
		t.Fatal("selection bounds", err)
	}
	in, err = Apply("toggle-switch", "toggle", Input{})
	if err != nil || !in.Enabled {
		t.Fatal("toggle", err)
	}
	if _, err := Apply("orbit-spinner", "next", Input{}); err == nil {
		t.Fatal("unexpected action")
	}
}

func TestGalleryEveryWidgetAndViewport(t *testing.T) {
	for _, d := range Catalog() {
		route := strings.ToLower(d.Category)
		for _, width := range []int{24, 40, 80, 120} {
			state := map[string]string{"state.pw_" + route: d.ID, "state.pw_motion": "Static"}
			v, bindings, animated, err := Gallery(route, width, time.Second, state)
			if err != nil || animated {
				t.Fatalf("%s: %v, animated=%t", d.ID, err, animated)
			}
			if !strings.Contains(bindings["preview.line1"], d.Title) && width >= 40 {
				t.Fatalf("%s: wrong preview", d.ID)
			}
			lines, spots, _ := v.RenderInteractive(bindings, width)
			if len(lines) == 0 || len(spots) == 0 {
				t.Fatalf("%s: missing real SML view or controls", d.ID)
			}
		}
	}
	if _, _, _, err := Gallery("unknown", 80, 0, nil); err == nil {
		t.Fatal("unknown gallery accepted")
	}
}
