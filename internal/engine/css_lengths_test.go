package engine

import "testing"

func TestCSSLengthExpressions(t *testing.T) {
	for _, tc := range []struct {
		expr  string
		width int
		want  int
	}{
		{"calc(100% - 4ch)", 80, 76},
		{"min(50%, 32ch)", 100, 32},
		{"max(20ch, 50%)", 30, 20},
		{"clamp(20ch, 70%, 60ch)", 100, 60},
		{"clamp(20ch, 70%, 60ch)", 24, 20},
		{"calc(min(50%, 40ch) + 4ch)", 100, 44},
		{"calc(50% + 50%)", 51, 51},
		{"calc(10ch - 30ch)", 80, -20},
	} {
		got, err := evaluateLengthExpression(tc.expr, tc.width)
		if err != nil || got != tc.want {
			t.Fatalf("%s at %d: got %d, %v; want %d", tc.expr, tc.width, got, err, tc.want)
		}
	}
}

func TestCSSLengthExpressionLayout(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		:root { --pane: clamp(20ch, 70%, 60ch); }
		box { width: var(--pane); }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><box align="center"><button id="open" label="Open" action="app.open"/></box></screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		width int
		want  int
	}{
		{12, 12}, {24, 20}, {60, 42}, {120, 60},
	} {
		frame, err := views["home"].RenderFrame(nil, tc.width)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range frame.Elements {
			if e.Kind == "box" {
				found = true
				if e.Bounds.Width != tc.want || e.Bounds.X != (tc.width-tc.want)/2 {
					t.Fatalf("incorrect box at %d: %+v", tc.width, e.Bounds)
				}
			}
		}
		if !found || len(frame.Hotspots) != 1 {
			t.Fatal("missing box or control")
		}
		spot := frame.Hotspots[0]
		if got, ok := frame.HitTest(spot.Start, spot.Line); !ok || got.ID != "open" {
			t.Fatal("incorrect responsive hit target")
		}
	}
}

func TestInvalidCSSLengthExpressions(t *testing.T) {
	for _, expr := range []string{
		"calc(10px + 5ch)", "calc(50% * 2)", "min(20ch)",
		"clamp(12ch, 100%)", "calc(101% - 2ch)", "calc(241ch)",
		"calc(50% +)", "min(1ch, 2ch, 3ch, 4ch, 5ch)",
		"calc(10ch) extra", "calc(calc(calc(calc(calc(1ch)))))",
		"calc(1ch / 0)", "calc(1ch + 2)",
	} {
		if err := validateLayoutStyleValue("width", expr); err == nil {
			t.Fatalf("accepted %s", expr)
		}
	}
	if err := validateLayoutStyleValue("gap", "calc(10ch)"); err == nil {
		t.Fatal("expression accepted on unsupported layout property")
	}
}
