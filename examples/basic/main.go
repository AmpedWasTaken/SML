package main

import (
	"fmt"
	"os"

	sml "github.com/AmpedWasTaken/SML"
)

func main() {
	document := []byte("<sml><panel title=\"SML · Quickstart\"><text>Reusable terminal interfaces, declarative layout, no runtime dependencies.</text><grid columns=\"2\"><metric label=\"Jobs\" value=\"42\"/><metric label=\"Healthy\" value=\"41\"/></grid></panel></sml>")
	root, err := sml.Parse(document)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := sml.Render(os.Stdout, root, sml.Options{Width: 72}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
