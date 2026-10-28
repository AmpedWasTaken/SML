// Command typed-monitor demonstrates a standalone SML host with typed data,
// atomic updates and a pure cell diff. It never connects to Sentinel.
package main

import (
	"fmt"
	"strings"

	sml "github.com/AmpedWasTaken/SML"
)

func main() {
	source := "<screen name=\"typed-monitor\" version=\"3\">" +
		"<panel title=\"Services\">" +
		"<text>Requests: {{metrics.requests}}</text>" +
		"<text>Healthy: {{metrics.healthy}}</text>" +
		"<button label=\"Refresh\" id=\"refresh\" action=\"monitor.refresh\"/>" +
		"</panel></screen>"
	view, err := sml.Parse([]byte(source))
	if err != nil {
		panic(err)
	}
	state, err := sml.NewBindingState(map[string]sml.Value{
		"metrics.requests": sml.IntValue(4),
		"metrics.healthy":  sml.BoolValue(true),
	})
	if err != nil {
		panic(err)
	}
	oldFrame, err := view.RenderTypedFrame(state.Snapshot().Values, 48)
	if err != nil {
		panic(err)
	}
	change, err := state.Apply(
		sml.BindingUpdate{Key: "metrics.requests", Value: sml.IntValue(5)},
		sml.BindingUpdate{Key: "metrics.healthy", Value: sml.BoolValue(false)},
	)
	if err != nil {
		panic(err)
	}
	nextFrame, err := view.RenderTypedFrame(state.Snapshot().Values, 48)
	if err != nil {
		panic(err)
	}
	diff, err := sml.DiffFrames(oldFrame, nextFrame)
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Join(nextFrame.Lines, "\n"))
	fmt.Printf("Revision %d: keys %v, changed runs %d, full redraw %t\n",
		change.Revision, change.ChangedKeys, len(diff.Runs), diff.FullRedraw)
	// This example intentionally does not dispatch button actions or write
	// ANSI patch commands; a real host owns auth and terminal output.
}
