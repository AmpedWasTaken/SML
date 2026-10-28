package sml_test

import (
	"reflect"
	"testing"

	sml "github.com/AmpedWasTaken/SML"
)

func TestPublicTypedBindingsAndBatchUpdates(t *testing.T) {
	view, err := sml.Parse([]byte("<screen name=\"external\" version=\"3\"><text>{{host.name}}: {{host.requests}}</text><button id=\"reload\" label=\"Reload\" action=\"app.reload\"/></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	store, err := sml.NewBindingState(map[string]sml.Value{
		"host.name":     sml.StringValue("server-one"),
		"host.requests": sml.IntValue(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	first := store.Snapshot()
	frame, err := view.RenderTypedFrame(first.Values, 80)
	if err != nil || len(frame.Hotspots) != 1 {
		t.Fatalf("typed facade render failed: %+v %v", frame, err)
	}
	change, err := store.Apply(sml.BindingUpdate{Key: "host.requests", Value: sml.IntValue(6)})
	if err != nil || change.Revision != 1 || !reflect.DeepEqual(change.ChangedKeys, []string{"host.requests"}) {
		t.Fatalf("public state update failed: %+v %v", change, err)
	}
	second := store.Snapshot()
	nextFrame, err := view.RenderTypedFrame(second.Values, 80)
	if err != nil {
		t.Fatal(err)
	}
	diff, err := sml.DiffFrames(frame, nextFrame)
	if err != nil || diff.FullRedraw || len(diff.Runs) == 0 {
		t.Fatalf("typed public data did not generate a frame diff: %+v %v", diff, err)
	}
	if value, ok := second.Values["host.requests"].Int64(); !ok || value != 6 {
		t.Fatal("typed state lost its integer representation")
	}
	var _ sml.ValueKind = sml.KindBool
	var _ sml.Value = sml.NullValue()
	var _ sml.BindingSnapshot = second
	var _ sml.BindingChange = change
	if err := sml.ValidateBindingValues(second.Values); err != nil {
		t.Fatal(err)
	}
	original := store.Snapshot()
	second.Values["host.requests"] = sml.IntValue(999)
	if store.Snapshot().Values["host.requests"] != original.Values["host.requests"] {
		t.Fatal("public snapshot leaked writable internal state")
	}
}
