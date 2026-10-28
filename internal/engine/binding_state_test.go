package engine

import (
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestBindingStateSnapshotsAreDetached(t *testing.T) {
	initial := map[string]Value{"service.status": StringValue("ready")}
	state, err := NewBindingState(initial)
	if err != nil {
		t.Fatal(err)
	}
	initial["service.status"] = StringValue("caller mutation")
	first := state.Snapshot()
	if first.Revision != 0 || first.Values["service.status"].String() != "ready" {
		t.Fatalf("constructor did not copy caller data: %+v", first)
	}
	first.Values["service.status"] = StringValue("snapshot mutation")
	second := state.Snapshot()
	if second.Values["service.status"].String() != "ready" {
		t.Fatalf("snapshot leaked a mutable state map: %+v", second)
	}
}

func TestBindingStateBatchIsAtomicAndIdempotent(t *testing.T) {
	state, err := NewBindingState(map[string]Value{
		"service.status": StringValue("idle"),
		"service.count":  IntValue(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := state.Apply(
		BindingUpdate{Key: "service.status", Value: StringValue("ready")},
		BindingUpdate{Key: "service.count", Value: IntValue(2)},
	)
	if err != nil || changed.Revision != 1 ||
		!reflect.DeepEqual(changed.ChangedKeys, []string{"service.count", "service.status"}) {
		t.Fatalf("unexpected atomic batch: %+v, %v", changed, err)
	}
	snapshot := state.Snapshot()
	if snapshot.Values["service.status"].String() != "ready" ||
		snapshot.Values["service.count"].String() != "2" {
		t.Fatalf("batch lost a value: %+v", snapshot)
	}
	same, err := state.Apply(
		BindingUpdate{Key: "service.count", Value: IntValue(2)},
		BindingUpdate{Key: "service.status", Value: StringValue("ready")},
	)
	if err != nil || same.Revision != 1 || len(same.ChangedKeys) != 0 {
		t.Fatalf("no-op update incremented revision: %+v %v", same, err)
	}
	for _, batch := range [][]BindingUpdate{
		{{Key: "service.status", Value: StringValue("bad")}, {Key: "derived.health", Value: StringValue("spoof")}},
		{{Key: "service.status", Value: StringValue("bad")}, {Key: "broken", Value: FloatValue(math.NaN())}},
		{{Key: "service.status", Value: StringValue("bad")}, {Key: "service.status", Value: StringValue("duplicate")}},
		{{Key: "service.status", Value: StringValue("bad")}, {Key: "service.count", Delete: true, Value: IntValue(1)}},
	} {
		if _, err := state.Apply(batch...); err == nil {
			t.Fatalf("accepted a partially invalid batch: %+v", batch)
		}
		unchanged := state.Snapshot()
		if unchanged.Revision != 1 || unchanged.Values["service.status"].String() != "ready" {
			t.Fatalf("failed batch caused a partial mutation: %+v", unchanged)
		}
	}
	deleted, err := state.Apply(BindingUpdate{Key: "service.count", Delete: true})
	if err != nil || deleted.Revision != 2 ||
		!reflect.DeepEqual(deleted.ChangedKeys, []string{"service.count"}) {
		t.Fatalf("failed to delete key: %+v, %v", deleted, err)
	}
	if _, exists := state.Snapshot().Values["service.count"]; exists {
		t.Fatal("deleted presentation key still exists")
	}
	again, err := state.Apply(BindingUpdate{Key: "service.count", Delete: true})
	if err != nil || again.Revision != 2 || len(again.ChangedKeys) != 0 {
		t.Fatalf("deleting missing key changed revision: %+v %v", again, err)
	}
}

func TestBindingStateRejectsOversizedAtomicBatch(t *testing.T) {
	state, err := NewBindingState(map[string]Value{
		"core.status": StringValue("ready"),
	})
	if err != nil {
		t.Fatal(err)
	}
	updates := make([]BindingUpdate, 0, 40)
	for i := 0; i < 40; i++ {
		updates = append(updates, BindingUpdate{
			Key:   fmt.Sprintf("extra.%d", i),
			Value: StringValue(string(make([]byte, 4000))),
		})
	}
	if _, err := state.Apply(updates...); err == nil {
		t.Fatal("accepted an oversized aggregate batch")
	}
	after := state.Snapshot()
	if len(after.Values) != 1 || after.Revision != 0 {
		t.Fatalf("aggregate batch partially applied: %+v", after)
	}
}

func TestBindingStateConcurrentWritersAndRenderSnapshots(t *testing.T) {
	state, err := NewBindingState(map[string]Value{
		"counter.one": IntValue(0),
		"counter.two": IntValue(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	view, err := Parse([]byte("<screen name=\"concurrent\"><field label=\"One\" key=\"counter.one\"/><field label=\"Two\" key=\"counter.two\"/></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "counter.one"
			if id%2 == 1 {
				key = "counter.two"
			}
			for i := 1; i <= 75; i++ {
				if _, err := state.Apply(BindingUpdate{Key: key, Value: IntValue(int64(id*1000 + i))}); err != nil {
					errs <- err
					return
				}
			}
		}(worker)
	}
	for reader := 0; reader < 4; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				snapshot := state.Snapshot()
				if _, err := view.RenderTypedFrame(snapshot.Values, 40); err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	final := state.Snapshot()
	if final.Revision == 0 || len(final.Values) != 2 {
		t.Fatalf("concurrent state lost data: %+v", final)
	}
}

func TestTypedStateRenderAndFrameDiff(t *testing.T) {
	state, err := NewBindingState(map[string]Value{
		"stats.requests": IntValue(8),
		"stats.online":   BoolValue(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	view, err := Parse([]byte("<screen name=\"metrics\"><text>Requests: {{stats.requests}}</text><text>Online: {{stats.online}}</text></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	oldFrame, err := view.RenderTypedFrame(state.Snapshot().Values, 40)
	if err != nil {
		t.Fatal(err)
	}
	change, err := state.Apply(BindingUpdate{Key: "stats.requests", Value: IntValue(9)})
	if err != nil || !reflect.DeepEqual(change.ChangedKeys, []string{"stats.requests"}) {
		t.Fatalf("unexpected changed keys: %+v %v", change, err)
	}
	newFrame, err := view.RenderTypedFrame(state.Snapshot().Values, 40)
	if err != nil {
		t.Fatal(err)
	}
	diff, err := DiffFrames(oldFrame, newFrame)
	if err != nil || diff.FullRedraw || len(diff.Runs) == 0 {
		t.Fatalf("typed state did not produce changed frame cells: %+v %v", diff, err)
	}
}
