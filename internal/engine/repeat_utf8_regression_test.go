package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// Repeated buttons with explicit IDs must remain individually actionable.
func TestRepeatExplicitControlIDsAreUniqueStableAndHitTestable(t *testing.T) {
	view, err := Parse([]byte(`<screen name="home" version="3">
		<repeat source="items" as="item">
			<button id="inspect" label="Inspect" page="home"/>
		</repeat>
	</screen>`))
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{"items": "first,second,third"}
	frame, err := view.RenderFrame(values, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 3 {
		t.Fatalf("expected three hotspots, got %d", len(frame.Hotspots))
	}
	seen := map[string]bool{}
	for _, spot := range frame.Hotspots {
		if seen[spot.ID] {
			t.Fatalf("repeated hotspot ID %q", spot.ID)
		}
		seen[spot.ID] = true
		for x := spot.Start; x < spot.End; x++ {
			if frame.Cells[spot.Line][x].ElementID != spot.ID {
				t.Fatalf("missing active cell at %d,%d", x, spot.Line)
			}
		}
	}
	again, err := view.RenderFrame(values, 80)
	if err != nil {
		t.Fatal(err)
	}
	for i, spot := range frame.Hotspots {
		if again.Hotspots[i].ID != spot.ID {
			t.Fatal("hotspot ID changed across redraws")
		}
	}
}

func TestLongInterpolationKeepsUTF8RuneBoundary(t *testing.T) {
	value := strings.Repeat("界", 1366)
	rendered := interpolate("{{item}}", map[string]string{"item": value})
	if !utf8.ValidString(rendered) || strings.ContainsRune(rendered, utf8.RuneError) {
		t.Fatalf("interpolation broke UTF-8: %q", rendered)
	}
	if len(rendered) != 4095 {
		t.Fatalf("expected 4095 UTF-8 bytes at a rune boundary, got %d", len(rendered))
	}
	if got := interpolate("{{item}}", map[string]string{"item": "🛠️ Ready"}); got != "🛠️ Ready" {
		t.Fatalf("short Unicode changed: %q", got)
	}
}

func TestNarrowPaddedPanelClipsHitboxInsteadOfBreakingFrame(t *testing.T) {
	view, err := Parse([]byte(`<screen name="home" version="3">
		<panel padding="3" border="round">
			<button id="narrow" label="Inspect" page="home"/>
		</panel>
	</screen>`))
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{12, 14, 16, 18, 24, 80} {
		frame, err := view.RenderFrame(nil, width)
		if err != nil {
			t.Fatalf("width %d: %v", width, err)
		}
		for _, spot := range frame.Hotspots {
			if spot.Start < 0 || spot.End > width || spot.Start >= spot.End {
				t.Fatalf("width %d invalid hotspot: %+v", width, spot)
			}
			if spot.Line < 0 || spot.Line >= frame.Height {
				t.Fatalf("width %d offscreen hotspot: %+v", width, spot)
			}
		}
		for _, line := range frame.Lines {
			if CellWidth(line) > width {
				t.Fatalf("width %d overflowing line: %q", width, line)
			}
		}
	}
}

func TestRepeatedComponentControlIDsStayUnique(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "components"), 0700); err != nil {
		t.Fatal(err)
	}
	component := `<component name="actions">
		<button id="inspect" label="Inspect" page="home"/>
	</component>`
	if err := os.WriteFile(filepath.Join(dir, "components", "actions.sml"), []byte(component), 0600); err != nil {
		t.Fatal(err)
	}
	view, err := Parse([]byte(`<screen name="home" version="3">
		<include name="actions"/><include name="actions"/>
	</screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.WithAssets(dir).RenderFrame(nil, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 2 || frame.Hotspots[0].ID == frame.Hotspots[1].ID {
		t.Fatalf("reused components share hotspot IDs: %#v", frame.Hotspots)
	}
}
