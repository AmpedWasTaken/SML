package engine

import (
	"strings"
	"testing"
)

func TestBoxedInputResponsiveGeometry(t *testing.T) {
	view, err := Parse([]byte(`<screen name="demo" version="3"><panel border="round" padding="1"><box width="40" align="center"><text align="center">Username</text><input id="username" key="state.username" label="Username" appearance="boxed" border="round" border-color="#76d8d7"/></box><button id="submit" label="Sign in" action="form.submit" align="center"/></panel></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{32, 60, 80, 120} {
		values := map[string]string{"state.username": "Alice"}
		frame, err := view.RenderFrame(values, width)
		if err != nil || len(frame.Hotspots) != 2 {
			t.Fatalf("width=%d: %v", width, err)
		}
		spot := frame.Hotspots[0]
		if spot.ID != "username" || spot.End-spot.Start != min(40, width-4)-2 {
			t.Fatalf("incorrect field geometry at %d: %+v", width, spot)
		}
		if got, ok := frame.HitTest(spot.Start, spot.Line); !ok || got.ID != "username" {
			t.Fatal("centered field lost hit target")
		}
		if !strings.Contains(frame.Lines[spot.Line], "Alice") || !strings.Contains(frame.Lines[spot.Line-1], "╭") {
			t.Fatal("input is not boxed")
		}
		painted, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true})
		if err != nil || !strings.Contains(strings.Join(painted, "\n"), "\x1b[38;2;118;216;215m") {
			t.Fatalf("missing field border color: %v", err)
		}
		plain, err := view.RenderChecked(values, width)
		if err != nil || !strings.Contains(strings.Join(plain, "\n"), "Alice") {
			t.Fatalf("legacy renderer lost boxed input: %v", err)
		}
	}
}

func TestInvalidBoxAndInputAppearance(t *testing.T) {
	for _, src := range []string{
		`<box width="999"><text>bad</text></box>`,
		`<box align="invalid"><text>bad</text></box>`,
		`<input key="state.value" label="Value" appearance="bad"/>`,
		`<input key="state.value" label="Value" appearance="boxed" border="none"/>`,
		`<input key="state.value" label="Value" appearance="boxed" border-color="bad"/>`,
	} {
		if _, err := Parse([]byte(`<screen name="demo" version="3">` + src + `</screen>`)); err == nil {
			t.Fatalf("accepted invalid markup: %s", src)
		}
	}
}
