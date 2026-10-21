// A small, host-neutral demonstration of action metadata.
// SML emits descriptors; the embedding application decides what to do with them.
package main

import (
	"fmt"
	"os"
	"strings"

	sml "github.com/AmpedWasTaken/SML"
)

func main() {
	view, err := sml.Parse([]byte(`<screen name="actions" version="3">
		<panel title="Service">
			<text>State: {{app.status}}</text>
			<button id="refresh" label="Refresh" action="app.refresh"/>
		</panel>
	</screen>`))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	frame, err := view.RenderFrame(map[string]string{"app.status": "online"}, 48)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(strings.Join(frame.Lines, "\n"))
	for _, hotspot := range frame.Hotspots {
		fmt.Printf("action=%s id=%s line=%d cells=%d..%d\n",
			hotspot.Action, hotspot.ID, hotspot.Line, hotspot.Start, hotspot.End)
	}
	// The host separately authenticates and authorizes the action before dispatch.
}
