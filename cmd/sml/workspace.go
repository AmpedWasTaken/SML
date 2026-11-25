package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	sml "github.com/AmpedWasTaken/SML"
)

// runWorkspace previews a validated asset revision without executing any
// action. Watch mode polls locally and leaves the previous preview visible
// when a designer saves an invalid intermediate file.
func runWorkspace(action string, args []string) int {
	fs := flag.NewFlagSet(action, flag.ContinueOnError)
	root := fs.String("assets", "assets", "asset root")
	width := fs.Int("width", 80, "viewport width (12..240)")
	height := fs.Int("height", 24, "viewport height (1..3000)")
	scroll := fs.Int("scroll", 0, "vertical viewport origin")
	ascii := fs.Bool("ascii", false, "conservative ASCII fallback")
	watch := fs.Bool("watch", false, "reload changed preview until interrupted")
	var raw bindings
	fs.Var(&raw, "set", "binding key=value (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *width < 12 || *width > 240 || *height < 1 || *height > 3000 || *scroll < 0 || (action == "debug" && *watch) {
		fmt.Fprintln(os.Stderr, "sml: invalid preview/debug arguments")
		return 2
	}
	name := fs.Arg(0)
	store, err := sml.NewAssetStore(*root, name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sml:", err)
		return 1
	}
	values := make(map[string]string, len(raw))
	for _, item := range raw {
		key, value, _ := strings.Cut(item, "=")
		values[key] = value
	}
	render := func() (sml.Frame, error) {
		view, _, _ := store.View(name)
		frame, err := view.RenderFrame(values, *width)
		if err != nil {
			return sml.Frame{}, err
		}
		frame, err = frame.Window(0, *scroll, *width, *height)
		if err == nil && *ascii {
			frame = frame.ASCII()
		}
		return frame, err
	}
	frame, err := render()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sml:", err)
		return 1
	}
	printFrame := func(frame sml.Frame) {
		for _, line := range frame.Lines {
			fmt.Println(line)
		}
	}
	printFrame(frame)
	if action == "debug" {
		fmt.Printf("Viewport: %dx%d, %d elements, %d hotspots\n", frame.Width, frame.Height, len(frame.Elements), len(frame.Hotspots))
		for _, element := range frame.Elements {
			r := element.Bounds
			fmt.Printf("%q <%s>: x=%d y=%d w=%d h=%d\n", element.ID, element.Kind, r.X, r.Y, r.Width, r.Height)
		}
		for _, spot := range frame.Hotspots {
			fmt.Printf("hotspot %q: row=%d start=%d end=%d\n", spot.ID, spot.Line, spot.Start, spot.End)
		}
		view, _, _ := store.View(name)
		deps := view.BindingDependencies()
		fmt.Printf("Binding keys: %s (conservative=%t)\n", strings.Join(deps.Keys, ", "), deps.Conservative)
	}
	if !*watch {
		return 0
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	lastError := ""
	for {
		select {
		case <-ctx.Done():
			return 0
		case <-ticker.C:
			_, err := store.Reload()
			var next sml.Frame
			if err == nil {
				next, err = render()
			}
			if err != nil {
				if err.Error() != lastError {
					fmt.Fprintln(os.Stderr, "sml reload:", err)
					lastError = err.Error()
				}
				continue
			}
			lastError = ""
			if !reflect.DeepEqual(frame, next) {
				fmt.Println("--- SML preview updated ---")
				printFrame(next)
				frame = next
			}
		}
	}
}
