package sml_test

import (
	"testing"

	sml "github.com/AmpedWasTaken/SML"
)

func TestPublicFrameDiffMatchesVisibleCells(t *testing.T) {
	view, err := sml.Parse([]byte(`<screen name="diff"><text>Health: {{health}}</text><button id="refresh" label="Refresh" action="app.refresh"/></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	before, err := view.RenderFrame(map[string]string{"health":"ready"}, 40)
	if err != nil {
		t.Fatal(err)
	}
	after, err := view.RenderFrame(map[string]string{"health":"online"}, 40)
	if err != nil {
		t.Fatal(err)
	}
	diff, err := sml.DiffFrames(before, after)
	if err != nil || diff.FullRedraw || len(diff.Runs) == 0 {
		t.Fatalf("expected minimal update, got %+v (%v)", diff, err)
	}
	var _ sml.FrameDiff = diff
	var _ sml.CellRun = diff.Runs[0]
	for _, run := range diff.Runs {
		if run.Row != 0 || run.Start < 0 || run.Start+len(run.Cells) > 40 {
			t.Fatalf("unexpected update coordinates %+v", run)
		}
	}
	if diff.HotspotsChanged {
		t.Fatal("label data changes must not remap a distinct button action")
	}
}
