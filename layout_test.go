package sml

import (
	"strings"
	"testing"
)

func TestFrameUsesGraphemeCells(t *testing.T) {
	view, err := Parse([]byte(`<screen name="wide"><text>界 👩‍💻 é</text><button id="refresh" label="Refresh" action="metrics.refresh"/></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 1 || frame.Hotspots[0].ID != "refresh" || frame.Hotspots[0].Action != "metrics.refresh" {
		t.Fatalf("%+v", frame.Hotspots)
	}
	for _, line := range frame.Lines {
		if CellWidth(line) > 40 {
			t.Fatalf("overflow %q", line)
		}
	}
	if !frame.Cells[0][1].Continuation {
		t.Fatal("wide rune didn't allocate two cells")
	}
}

func TestFrameStructuralGeometryAndDuplicateLabels(t *testing.T) {
	view, err := Parse([]byte(`<screen name="two" version="3">
      <row gap="2" collapse-below="40">
        <panel title="Left"><button id="left" label="Open" action="app.left"/></panel>
        <panel title="Right"><button id="right" label="Open" action="app.right"/></panel>
      </row>
    </screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 2 {
		t.Fatalf("spots=%+v, lines=%v", frame.Hotspots, frame.Lines)
	}
	left, right := frame.Hotspots[0], frame.Hotspots[1]
	if left.Line != right.Line || left.Start >= right.Start || left.Action == right.Action {
		t.Fatalf("incorrect geometry: %+v %+v", left, right)
	}
	if frame.Cells[left.Line][left.Start].ElementID != "left" || frame.Cells[right.Line][right.Start].ElementID != "right" {
		t.Fatal("control cell attribution missing")
	}
}

func TestFrameConditionalAndRepeatedIds(t *testing.T) {
	src := `<screen name="conditional" version="3">
      <if field="is.ready" equals="true">
        <button label="Same" action="active.check"/>
        <otherwise><button label="Same" action="inactive.check"/></otherwise>
      </if>
      <repeat source="items" as="item" limit="3">
        <button label="Next" action="next.page"/>
      </repeat>
    </screen>`
	view, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(map[string]string{"is.ready": "false", "items": "a,b,c"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Hotspots) != 4 {
		t.Fatalf("wrong count %+v", frame.Hotspots)
	}
	if frame.Hotspots[0].Action != "inactive.check" {
		t.Fatal(frame.Hotspots)
	}
	ids := map[string]bool{}
	for _, spot := range frame.Hotspots {
		if ids[spot.ID] {
			t.Fatalf("repeat collision %+v", frame.Hotspots)
		}
		ids[spot.ID] = true
	}
}

func TestFrameTextWrappingAndNarrowLayout(t *testing.T) {
	view, err := Parse([]byte(`<screen name="wrap"><text>These are several very wide words including 界界 and emoji 👩‍💻 here</text></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	f, err := view.RenderFrame(nil, 18)
	if err != nil {
		t.Fatal(err)
	}
	if f.Height < 3 {
		t.Fatalf("did not wrap: %v", f.Lines)
	}
	for _, line := range f.Lines {
		if CellWidth(line) > 18 {
			t.Fatalf("line exceeds cells: %q", line)
		}
	}
	if !strings.Contains(strings.Join(f.Lines, ""), "👩‍💻") {
		t.Fatal(f.Lines)
	}
}
